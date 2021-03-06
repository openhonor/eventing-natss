package source

import (
	"fmt"
	"github.com/nats-io/stan.go"
	"testing"
)

var (
	clusterID = "nats://10.224.201.88:30180"
	clientID  = "$ip"
)

func TestNatssClient(t *testing.T) {
	var (
		err error
	)
	sc, _ := stan.Connect(clusterID, clientID)
	// Simple Synchronous Publisher
	err = sc.Publish("foo", []byte("Hello World")) // does not return until an ack has been received from NATS Streaming

	// Simple Async Subscriber
	sub, err := sc.QueueSubscribe("foo", "foo_group", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
		_ = m.Ack()

	}, stan.SetManualAckMode())
	if err != nil {
		panic(err)
	}
	defer func() {
		if sub != nil {
			err = sub.Close()
		}
	}()

}
