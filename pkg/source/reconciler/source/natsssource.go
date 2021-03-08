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

package source

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing/pkg/utils"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"knative.dev/eventing-natss/pkg/apis/sources/v1beta1"
	"knative.dev/eventing-natss/pkg/source/reconciler/source/resources"

	"k8s.io/client-go/kubernetes"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	"knative.dev/eventing-natss/pkg/client/clientset/versioned"
	reconcilernatsssource "knative.dev/eventing-natss/pkg/client/injection/reconciler/sources/v1beta1/natsssource"
	listers "knative.dev/eventing-natss/pkg/client/listers/sources/v1beta1"
)

const (
	raImageEnvVar                = "NATSS_RA_IMAGE"
	natssSourceDeploymentCreated = "NatssSourceDeploymentCreated"
	natssSourceDeploymentUpdated = "NatssSourceDeploymentUpdated"
	natssSourceDeploymentScaled  = "NatssSourceDeploymentScaled"
	natssSourceDeploymentFailed  = "NatssSourceDeploymentFailed"
	natssSourceDeploymentDeleted = "NatssSourceDeploymentDeleted"
	component                    = "natsssource"
)

// newDeploymentCreated makes a new reconciler event with event type Normal, and
// reason NatssSourceDeploymentCreated.
func newDeploymentCreated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, natssSourceDeploymentCreated, "NatssSource created deployment: \"%s/%s\"", namespace, name)
}

// deploymentUpdated makes a new reconciler event with event type Normal, and
// reason NatssSourceDeploymentUpdated.
func deploymentUpdated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, natssSourceDeploymentUpdated, "NatssSource updated deployment: \"%s/%s\"", namespace, name)
}

// deploymentScaled makes a new reconciler event with event type Normal, and
// reason NatssSourceDeploymentScaled
func deploymentScaled(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, natssSourceDeploymentScaled, "NatssSource scaled deployment: \"%s/%s\"", namespace, name)
}

// newDeploymentFailed makes a new reconciler event with event type Warning, and
// reason NatssSourceDeploymentFailed.
func newDeploymentFailed(namespace, name string, err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, natssSourceDeploymentFailed, "NatssSource failed to create deployment: \"%s/%s\", %w", namespace, name, err)
}

type Reconciler struct {
	// KubeClientSet allows us to talk to the k8s for core APIs
	KubeClientSet kubernetes.Interface

	receiveAdapterImage string

	natssLister      listers.NatssSourceLister
	deploymentLister appsv1listers.DeploymentLister

	natssClientSet versioned.Interface
	loggingContext context.Context

	sinkResolver *resolver.URIResolver

	configs NatssSourceConfigAccessor
}

// Check that our Reconciler implements Interface
var _ reconcilernatsssource.Interface = (*Reconciler)(nil)

func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1beta1.NatssSource) pkgreconciler.Event {
	src.Status.InitializeConditions()

	if (src.Spec.Sink == duckv1.Destination{}) {
		src.Status.MarkNoSink("SinkMissing", "")
		return fmt.Errorf("spec.sink missing")
	}

	dest := src.Spec.Sink.DeepCopy()
	if dest.Ref != nil {
		// To call URIFromDestination(), dest.Ref must have a Namespace. If there is
		// no Namespace defined in dest.Ref, we will use the Namespace of the source
		// as the Namespace of dest.Ref.
		if dest.Ref.Namespace == "" {
			dest.Ref.Namespace = src.GetNamespace()
		}
	}
	sinkURI, err := r.sinkResolver.URIFromDestinationV1(ctx, *dest, src)
	if err != nil {
		src.Status.MarkNoSink("NotFound", "")
		//delete adapter deployment if sink not found
		if err := r.deleteReceiveAdapter(ctx, src); err != nil && !apierrors.IsNotFound(err) {
			logging.FromContext(ctx).Error("Unable to delete receiver adapter when sink is missing", zap.Error(err))
		}
		return fmt.Errorf("getting sink URI: %v", err)
	}
	src.Status.MarkSink(sinkURI)

	if val, ok := src.GetLabels()[v1beta1.NatssKeyTypeLabel]; ok {
		found := false
		for _, allowed := range v1beta1.NatssKeyTypeAllowed {
			if allowed == val {
				found = true
			}
		}
		if !found {
			src.Status.MarkKeyTypeIncorrect("IncorrectNatssKeyTypeLabel", "Invalid value for %s: %s. Allowed: %v", v1beta1.NatssKeyTypeLabel, val, v1beta1.NatssKeyTypeAllowed)
			logging.FromContext(ctx).Errorf("Invalid value for %s: %s. Allowed: %v", v1beta1.NatssKeyTypeLabel, val, v1beta1.NatssKeyTypeAllowed)
			return errors.New("IncorrectNatssKeyTypeLabel")
		} else {
			src.Status.MarkKeyTypeCorrect()
		}
	}

	// TODO(mattmoor): create NatssBinding for the receive adapter.

	ra, err := r.createReceiveAdapter(ctx, src, sinkURI)
	if err != nil {
		var event *pkgreconciler.ReconcilerEvent
		isReconcilerEvent := pkgreconciler.EventAs(err, &event)
		if isReconcilerEvent && event.EventType != corev1.EventTypeNormal {
			logging.FromContext(ctx).Error("Unable to create the receive adapter. Reconciler error", zap.Error(err))
			return err
		} else if !isReconcilerEvent {
			logging.FromContext(ctx).Error("Unable to create the receive adapter. Generic error", zap.Error(err))
			return err
		}
	}
	src.Status.MarkDeployed(ra)
	src.Status.CloudEventAttributes = r.createCloudEventAttributes(src)

	return nil
}

func (r *Reconciler) createReceiveAdapter(ctx context.Context, src *v1beta1.NatssSource, sinkURI *apis.URL) (*appsv1.Deployment, error) {
	raArgs := resources.ReceiveAdapterArgs{
		Image:          r.receiveAdapterImage,
		Source:         src,
		Labels:         resources.GetLabels(src.Name),
		SinkURI:        sinkURI.String(),
		AdditionalEnvs: r.configs.ToEnvVars(),
	}
	expected := resources.MakeReceiveAdapter(&raArgs)

	ra, err := r.KubeClientSet.AppsV1().Deployments(src.Namespace).Get(ctx, expected.Name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		// Issue eventing#2842: Adater deployment name uses kmeta.ChildName. If a deployment by the previous name pattern is found, it should
		// be deleted. This might cause temporary downtime.
		if deprecatedName := utils.GenerateFixedName(raArgs.Source, fmt.Sprintf("natsssource-%s", raArgs.Source.Name)); deprecatedName != expected.Name {
			if err := r.KubeClientSet.AppsV1().Deployments(src.Namespace).Delete(ctx, deprecatedName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("error deleting deprecated named deployment: %v", err)
			}
			controller.GetEventRecorder(ctx).Eventf(src, corev1.EventTypeNormal, natssSourceDeploymentDeleted, "Deprecated deployment removed: \"%s/%s\"", src.Namespace, deprecatedName)
		}
		ra, err = r.KubeClientSet.AppsV1().Deployments(src.Namespace).Create(ctx, expected, metav1.CreateOptions{})
		if err != nil {
			return nil, newDeploymentFailed(ra.Namespace, ra.Name, err)
		}
		return ra, newDeploymentCreated(ra.Namespace, ra.Name)
	} else if err != nil {
		logging.FromContext(ctx).Error("Unable to get an existing receive adapter", zap.Error(err))
		return nil, err
	} else if !metav1.IsControlledBy(ra, src) {
		return nil, fmt.Errorf("deployment %q is not owned by NatssSource %q", ra.Name, src.Name)
	} else if podSpecChanged(ra.Spec.Template.Spec, expected.Spec.Template.Spec) {
		ra.Spec.Template.Spec = expected.Spec.Template.Spec
		if ra, err = r.KubeClientSet.AppsV1().Deployments(src.Namespace).Update(ctx, ra, metav1.UpdateOptions{}); err != nil {
			return ra, err
		}
		return ra, deploymentUpdated(ra.Namespace, ra.Name)
	} else if deref(ra.Spec.Replicas) != deref(expected.Spec.Replicas) {
		ra.Spec.Replicas = expected.Spec.Replicas
		if ra, err = r.KubeClientSet.AppsV1().Deployments(src.Namespace).Update(ctx, ra, metav1.UpdateOptions{}); err != nil {
			return ra, err
		}
		return ra, deploymentScaled(ra.Namespace, ra.Name)
	} else {
		logging.FromContext(ctx).Debug("Reusing existing receive adapter", zap.Any("receiveAdapter", ra))
	}
	return ra, nil
}

//deleteReceiveAdapter deletes the receiver adapter deployment if any
func (r *Reconciler) deleteReceiveAdapter(ctx context.Context, src *v1beta1.NatssSource) error {
	name := kmeta.ChildName(fmt.Sprintf("natsssource-%s-", src.Name), string(src.GetUID()))

	return r.KubeClientSet.AppsV1().Deployments(src.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func podSpecChanged(oldPodSpec corev1.PodSpec, newPodSpec corev1.PodSpec) bool {
	if !equality.Semantic.DeepDerivative(newPodSpec, oldPodSpec) {
		return true
	}
	if len(oldPodSpec.Containers) != len(newPodSpec.Containers) {
		return true
	}
	for i := range newPodSpec.Containers {
		if !equality.Semantic.DeepEqual(newPodSpec.Containers[i].Env, oldPodSpec.Containers[i].Env) {
			return true
		}
	}
	return false
}

func (r *Reconciler) createCloudEventAttributes(src *v1beta1.NatssSource) []duckv1.CloudEventAttributes {
	ceAttributes := make([]duckv1.CloudEventAttributes, 0, len(src.Spec.Subjects))
	for i := range src.Spec.Subjects {
		topics := strings.Split(src.Spec.Subjects[i], ",")
		for _, topic := range topics {
			ceAttributes = append(ceAttributes, duckv1.CloudEventAttributes{
				Type:   v1beta1.NatssEventType,
				Source: v1beta1.NatssEventSource(src.Namespace, src.Name, topic),
			})
		}
	}
	return ceAttributes
}

func deref(i *int32) int32 {
	if i == nil {
		return 1
	}
	return *i
}
