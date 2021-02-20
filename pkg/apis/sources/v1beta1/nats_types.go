package v1beta1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/eventing-natss/pkg/apis/duck/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// +genclient
// +genclient:method=GetScale,verb=get,subresource=scale,result=k8s.io/api/autoscaling/v1.Scale
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NatssSource is the Schema for the natsssources API.
// +k8s:openapi-gen=true
type NatssSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatssSourceSpec   `json:"spec,omitempty"`
	Status NatssSourceStatus `json:"status,omitempty"`
}

type NatssSourceSpec struct {
	// Number of desired consumers running in the consumer group. Defaults to 1.
	//
	// This is a pointer to distinguish between explicit
	// zero and not specified.
	// +optional
	Consumers *int32 `json:"consumers,omitempty"`

	NatssAuthSpec `json:",inline"`

	// ConsumerGroupID is the consumer group ID.
	// +optional
	ConsumerGroup string `json:"consumerGroup,omitempty"`

	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`
}

type NatssAuthSpec struct {
	Servers []*NatssBootstrapServerSpec `json:"servers"`
}

type NatssBootstrapServerSpec struct {
	NatssUrl  string `json:"natss_url"`
	ClusterID string `json:"cluster_id"`
	ClientID  string `json:"client_id,omitempty"`
}

// NatssSourceStatus defines the observed state of NatssSource.
type NatssSourceStatus struct {
	// inherits duck/v1 SourceStatus, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last
	//   processed by the controller.
	// * Conditions - the latest available observations of a resource's current
	//   state.
	// * SinkURI - the current active sink URI that has been configured for the
	//   Source.
	duckv1.SourceStatus `json:",inline"`

	// Total number of consumers actually running in the consumer group.
	// +optional
	Consumers int32 `json:"consumers,omitempty"`

	// Implement Placeable.
	// +optional
	v1alpha1.Placeable `json:",inline"`
}

const (
	NatssEventType = "dev.knative.natss.event"
)

func NatssEventSource(namespace, natssSourceName, subject string) string {
	return fmt.Sprintf("/apis/v1/namespaces/%s/natsssources/%s#%s", namespace, natssSourceName, subject)
}
