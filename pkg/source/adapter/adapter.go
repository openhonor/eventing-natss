/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package natss

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/nats-io/stan.go"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"knative.dev/eventing-natss/pkg/common/consumer"
	"knative.dev/eventing-natss/pkg/source"
	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/logging"
	pkgsource "knative.dev/pkg/source"
)

const (
	resourceGroup = "natsssources.sources.knative.dev"
)

type AdapterConfig struct {
	adapter.EnvConfig
	Subjects []string `envconfig:"NATSS_SUBJECTS" required:"true"`

	Name string `envconfig:"NAME" required:"true"`
}

func NewEnvConfig() adapter.EnvConfigAccessor {
	return &AdapterConfig{}
}

type Adapter struct {
	config            *AdapterConfig
	httpMessageSender *kncloudevents.HTTPMessageSender
	reporter          pkgsource.StatsReporter
	logger            *zap.SugaredLogger
	rateLimiter       *rate.Limiter
}

var _ adapter.MessageAdapter = (*Adapter)(nil)
var _ adapter.MessageAdapterConstructor = NewAdapter

func NewAdapter(ctx context.Context, processed adapter.EnvConfigAccessor, httpMessageSender *kncloudevents.HTTPMessageSender, reporter pkgsource.StatsReporter) adapter.MessageAdapter {
	logger := logging.FromContext(ctx)
	config := processed.(*AdapterConfig)

	return &Adapter{
		config:            config,
		httpMessageSender: httpMessageSender,
		reporter:          reporter,
		logger:            logger,
	}
}

func (a *Adapter) Start(ctx context.Context) error {
	return a.start(ctx.Done())
}

func (a *Adapter) start(stopCh <-chan struct{}) error {
	a.logger.Infow("Starting with config: ",
		zap.String("Subjects", strings.Join(a.config.Subjects, ",")),
		zap.String("SinkURI", a.config.Sink),
		zap.String("Name", a.config.Name),
		zap.String("Namespace", a.config.Namespace),
	)

	// init consumer connect
	natsEnv, _, err := source.NewConfig(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create the config: %w", err)
	}

	if natsEnv.ClientID == "" {
		natsEnv.ClientID = fmt.Sprintf("%s-%s", a.config.Namespace, a.config.Name)
	}
	consumerGroupFactory := consumer.NewConsumerGroupFactory(natsEnv.NatssUrl, natsEnv.ClusterID, natsEnv.ClientID)
	err = consumerGroupFactory.StartConsumerGroup(a.config.Subjects, a.logger, a)
	if err != nil {
		panic(err)
	}

	<-stopCh
	a.logger.Info("Shutting down...")
	return nil
}

func (a *Adapter) SetReady(_ bool) {}

func (a *Adapter) Handle(ctx context.Context, msg *stan.Msg) (bool, error) {
	if a.rateLimiter != nil {
		a.rateLimiter.Wait(ctx)
	}

	ctx, span := trace.StartSpan(ctx, "natss-source")
	defer span.End()

	req, err := a.httpMessageSender.NewCloudEventRequest(ctx)
	if err != nil {
		return false, err
	}

	err = a.ConsumerMessageToHttpRequest(ctx, span, msg, req)
	if err != nil {
		a.logger.Debug("failed to create request", zap.Error(err))
		return true, err
	}

	res, err := a.httpMessageSender.Send(req)

	if err != nil {
		a.logger.Debug("Error while sending the message", zap.Error(err))
		return false, err // Error while sending, don't commit offset
	}

	// Always try to read and close body so the connection can be reused afterwards
	if res.Body != nil {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}

	if res.StatusCode/100 != 2 {
		a.logger.Debug("Unexpected status code", zap.Int("status code", res.StatusCode))
		return false, fmt.Errorf("%d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	reportArgs := &pkgsource.ReportArgs{
		Namespace:     a.config.Namespace,
		Name:          a.config.Name,
		ResourceGroup: resourceGroup,
	}

	_ = a.reporter.ReportEventCount(reportArgs, res.StatusCode)
	return true, nil
}
