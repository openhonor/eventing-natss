package natss

import (
	"context"
	natsscloudevents "github.com/cloudevents/sdk-go/protocol/stan/v2"
	"github.com/cloudevents/sdk-go/v2/extensions"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/nats-io/stan.go"
	"go.opencensus.io/trace"
	nethttp "net/http"
)

func (a *Adapter) ConsumerMessageToHttpRequest(ctx context.Context, span *trace.Span, msg *stan.Msg, req *nethttp.Request) error {
	message, err := natsscloudevents.NewMessage(msg, natsscloudevents.WithManualAcks())
	if err != nil {
		return err
	}
	// Build tracing ext to write it as output
	tracingExt := extensions.FromSpanContext(span.SpanContext())

	a.logger.Debug("Message is not a CloudEvent -> We need to translate it to a valid CloudEvent")

	return http.WriteRequest(ctx, message, req, tracingExt.WriteTransformer())
}
