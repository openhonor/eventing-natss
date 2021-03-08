/*
Copyright 2020 The Knative Authors

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

package v1beta1

import (
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
)

const (
	// NatssConditionReady has status True when the NatssSource is ready to send events.
	NatssConditionReady = apis.ConditionReady

	// NatssConditionSinkProvided has status True when the NatssSource has been configured with a sink target.
	NatssConditionSinkProvided apis.ConditionType = "SinkProvided"

	// NatssConditionDeployed has status True when the NatssSource has had it's receive adapter deployment created.
	NatssConditionDeployed apis.ConditionType = "Deployed"

	// NatssConditionKeyType is True when the NatssSource has been configured with valid key type for
	// the key deserializer.
	NatssConditionKeyType apis.ConditionType = "KeyTypeCorrect"
)

var (
	NatssSourceCondSet = apis.NewLivingConditionSet(
		NatssConditionSinkProvided,
		NatssConditionDeployed)

	NatssCondSetLock = sync.RWMutex{}
)

// RegisterAlternateNatssConditionSet register an alternate apis.ConditionSet.
func RegisterAlternateNatssConditionSet(conditionSet apis.ConditionSet) {
	NatssCondSetLock.Lock()
	defer NatssCondSetLock.Unlock()

	NatssSourceCondSet = conditionSet
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*NatssSource) GetConditionSet() apis.ConditionSet {
	return NatssSourceCondSet
}

func (s *NatssSourceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return NatssSourceCondSet.Manage(s).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (s *NatssSourceStatus) IsReady() bool {
	return NatssSourceCondSet.Manage(s).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *NatssSourceStatus) InitializeConditions() {
	NatssSourceCondSet.Manage(s).InitializeConditions()
}

// MarkSink sets the condition that the source has a sink configured.
func (s *NatssSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if !uri.IsEmpty() {
		NatssSourceCondSet.Manage(s).MarkTrue(NatssConditionSinkProvided)
	} else {
		NatssSourceCondSet.Manage(s).MarkUnknown(NatssConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *NatssSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	NatssSourceCondSet.Manage(s).MarkFalse(NatssConditionSinkProvided, reason, messageFormat, messageA...)
}

func DeploymentIsAvailable(d *appsv1.DeploymentStatus, def bool) bool {
	// Check if the Deployment is available.
	for _, cond := range d.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			return cond.Status == "True"
		}
	}
	return def
}

// MarkDeployed sets the condition that the source has been deployed.
func (s *NatssSourceStatus) MarkDeployed(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		NatssSourceCondSet.Manage(s).MarkTrue(NatssConditionDeployed)

		// Propagate the number of consumers
		s.Consumers = d.Status.Replicas
	} else {
		// I don't know how to propagate the status well, so just give the name of the Deployment
		// for now.
		NatssSourceCondSet.Manage(s).MarkFalse(NatssConditionDeployed, "DeploymentUnavailable", "The Deployment '%s' is unavailable.", d.Name)
	}
}

// MarkDeploying sets the condition that the source is deploying.
func (s *NatssSourceStatus) MarkDeploying(reason, messageFormat string, messageA ...interface{}) {
	NatssSourceCondSet.Manage(s).MarkUnknown(NatssConditionDeployed, reason, messageFormat, messageA...)
}

// MarkNotDeployed sets the condition that the source has not been deployed.
func (s *NatssSourceStatus) MarkNotDeployed(reason, messageFormat string, messageA ...interface{}) {
	NatssSourceCondSet.Manage(s).MarkFalse(NatssConditionDeployed, reason, messageFormat, messageA...)
}

func (s *NatssSourceStatus) MarkKeyTypeCorrect() {
	NatssSourceCondSet.Manage(s).MarkTrue(NatssConditionKeyType)
}

func (s *NatssSourceStatus) MarkKeyTypeIncorrect(reason, messageFormat string, messageA ...interface{}) {
	NatssSourceCondSet.Manage(s).MarkFalse(NatssConditionKeyType, reason, messageFormat, messageA...)
}
