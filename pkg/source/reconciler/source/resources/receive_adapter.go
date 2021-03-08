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

package resources

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing-natss/pkg/apis/sources/v1beta1"
	"knative.dev/pkg/kmeta"
)

type ReceiveAdapterArgs struct {
	Image          string
	Source         *v1beta1.NatssSource
	Labels         map[string]string
	SinkURI        string
	AdditionalEnvs []corev1.EnvVar
}

func MakeReceiveAdapter(args *ReceiveAdapterArgs) *v1.Deployment {
	env := append([]corev1.EnvVar{{
		Name:  "NATSS_URL",
		Value: args.Source.Spec.NatssURL,
	}, {
		Name:  "NATSS_CLUSTER_ID",
		Value: args.Source.Spec.ClusterID,
	}, {
		Name: "NATSS_CLIENT_ID",
		Value: args.Source.Spec.ClientID,
	}, {
		Name:  "NATSS_SUBJECTS",
		Value: strings.Join(args.Source.Spec.Subjects, ","),
	}, {
		Name:  "NATSS_CONSUMER_GROUP",
		Value: args.Source.Spec.ConsumerGroup,
	}, {
		Name:  "K_SINK",
		Value: args.SinkURI,
	}, {
		Name:  "NAME",
		Value: args.Source.Name,
	}, {
		Name:  "NAMESPACE",
		Value: args.Source.Namespace,
	}}, args.AdditionalEnvs...)

	if val, ok := args.Source.GetLabels()[v1beta1.NatssKeyTypeLabel]; ok {
		env = append(env, corev1.EnvVar{
			Name:  "KEY_TYPE",
			Value: val,
		})
	}

	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kmeta.ChildName(fmt.Sprintf("natsssource-%s-", args.Source.Name), string(args.Source.GetUID())),
			Namespace: args.Source.Namespace,
			Labels:    args.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(args.Source),
			},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: args.Labels,
			},
			Replicas: args.Source.Spec.Consumers,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
					Labels: args.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "receive-adapter",
							Image: args.Image,
							Env:   env,
							Ports: []corev1.ContainerPort{
								{Name: "metrics", ContainerPort: 9090},
								{Name: "profiling", ContainerPort: 8008},
							},
						},
					},
				},
			},
		},
	}
}

// appendEnvFromSecretKeyRef returns env with an EnvVar appended
// setting key to the secret and key described by ref.
// If ref is nil, env is returned unchanged.
func appendEnvFromSecretKeyRef(env []corev1.EnvVar, key string, ref *corev1.SecretKeySelector) []corev1.EnvVar {
	if ref == nil {
		return env
	}

	env = append(env, corev1.EnvVar{
		Name: key,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: ref,
		},
	})

	return env
}
