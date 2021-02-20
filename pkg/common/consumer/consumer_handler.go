package consumer

import (
	"context"
	"github.com/nats-io/stan.go"
)

type NatssConsumerHandler interface {
	// When this function returns true, the consumer group offset is marked as consumed.
	// The returned error is enqueued in errors channel.
	Handle(context context.Context, message *stan.Msg) (bool, error)
	SetReady(ready bool)
}
