package consumer

import (
	"context"
	"fmt"
	natsscloudevents "github.com/cloudevents/sdk-go/protocol/stan/v2"
	"go.uber.org/zap"
	"knative.dev/eventing-natss/pkg/stanutil"
	"strings"
	"time"

	"github.com/nats-io/stan.go"
)

type NatssConsumerGroupFactory interface {
	StartConsumerGroup(subjects []string, logger *zap.SugaredLogger, handler NatssConsumerHandler) error
}

type natssConsumerGroupFactoryImpl struct {
	addr      string
	clusterID string
	clientID  string
}

func NewConsumerGroupFactory(addr string, clusterID string, clientID string) NatssConsumerGroupFactory {
	return &natssConsumerGroupFactoryImpl{addr: addr, clusterID: clusterID, clientID: clientID}
}

var _ NatssConsumerGroupFactory = (*natssConsumerGroupFactoryImpl)(nil)

func (n *natssConsumerGroupFactoryImpl) StartConsumerGroup(subjects []string, logger *zap.SugaredLogger, handler NatssConsumerHandler) error {
	sc, err := stanutil.Connect(n.clusterID, n.clientID, n.addr, logger)
	if err != nil {
		return fmt.Errorf("connect natss %s,with clusterID: %s failed, err: %s", n.addr, n.clusterID, err.Error())
	}
	mcb := func(stanMsg *stan.Msg) {
		defer func() {
			if r := recover(); r != nil {
				logger.Warn("Panic happened while handling a message",
					zap.String("messages", stanMsg.String()),
					zap.String("subjects", strings.Join(subjects, ",")),
					zap.Any("panic value", r),
				)
			}
		}()
		_, err := handler.Handle(context.TODO(), stanMsg)
		if err != nil {
			logger.Error("failed to handle message", zap.Error(err))
		}

		logger.Debug("Dispatch details")
		if err := stanMsg.Ack(); err != nil {
			logger.Error("failed to acknowledge message", zap.Error(err))
		}

		logger.Debug("message dispatched")
	}
	subscriber := &natsscloudevents.RegularSubscriber{}
	_, err = subscriber.Subscribe(*sc, subjects[0], mcb, stan.SetManualAckMode(), stan.AckWait(1*time.Minute))
	// TODO  unsubscribe

	return err
}
