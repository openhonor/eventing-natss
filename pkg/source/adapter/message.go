package natss

import (
	"context"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	natsscloudevents "github.com/cloudevents/sdk-go/protocol/stan/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/extensions"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/nats-io/stan.go"
	"go.opencensus.io/trace"

	"knative.dev/eventing-natss/pkg/apis/sources/v1beta1"
)

func (a *Adapter) ConsumerMessageToHttpRequest(ctx context.Context, span *trace.Span, stanMsg *stan.Msg, req *nethttp.Request) error {
	msg, err := natsscloudevents.NewMessage(stanMsg, natsscloudevents.WithManualAcks())
	if err != nil {
		return err
	}
	// Build tracing ext to write it as output
	tracingExt := extensions.FromSpanContext(span.SpanContext())

	a.logger.Debug("Message is not a CloudEvent -> We need to translate it to a valid CloudEvent")

	event := cloudevents.NewEvent()
	event.SetID(makeEventId(msg.Msg.Subject, msg.Msg.Sequence))
	event.SetTime(time.Unix(msg.Msg.Timestamp/1000000000, 0))
	event.SetType(v1beta1.NatssEventType)
	event.SetSource(v1beta1.NatssEventSource(a.config.Namespace, a.config.Name, msg.Msg.Subject))
	event.SetSubject(msg.Msg.Subject)
	err = event.SetData(cloudevents.ApplicationJSON, msg.Msg.Data)
	if err != nil {
		return err
	}

	return http.WriteRequest(ctx, binding.ToMessage(&event), req, tracingExt.WriteTransformer())
}

func makeEventId(subject string, sequence uint64) string {
	var str strings.Builder
	str.WriteString("subject:")
	str.WriteString(subject)
	str.WriteByte('#')
	str.WriteString("sequence:")
	str.WriteString(strconv.Itoa(int(sequence)))
	return str.String()
}
