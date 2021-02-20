package consumer

import (
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

type NatssConsumerGroupFactory interface {
	StartConsumerGroup(groupID string, subjects []string, logger *zap.SugaredLogger, handler NatssConsumerHandler) error
}

type natssConsumerGroupFactoryImpl struct {
	clusterID string
	clientID  string
}

func (n *natssConsumerGroupFactoryImpl) StartConsumerGroup(groupID string, subjects []string, logger *zap.SugaredLogger, handler NatssConsumerHandler) error {
	sc, _ := stan.Connect(n.clusterID, n.clientID)
	logger.Error("abc %v", sc)
	return nil
}
