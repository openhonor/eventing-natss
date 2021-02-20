package v1beta1

import (
	"sync"

	"knative.dev/pkg/apis"
)

const (
	// KafkaConditionReady has status True when the KafkaSource is ready to send events.
	KafkaConditionReady = apis.ConditionReady

	// KafkaConditionSinkProvided has status True when the KafkaSource has been configured with a sink target.
	KafkaConditionSinkProvided apis.ConditionType = "SinkProvided"

	// KafkaConditionDeployed has status True when the KafkaSource has had it's receive adapter deployment created.
	KafkaConditionDeployed apis.ConditionType = "Deployed"

	// KafkaConditionKeyType is True when the KafkaSource has been configured with valid key type for
	// the key deserializer.
	KafkaConditionKeyType apis.ConditionType = "KeyTypeCorrect"
)

var (
	KafkaSourceCondSet = apis.NewLivingConditionSet(
		KafkaConditionSinkProvided,
		KafkaConditionDeployed)

	kafkaCondSetLock = sync.RWMutex{}
)


