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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/eventing-natss/pkg/apis/sources/v1beta1"
	sourcessv1beta1 "knative.dev/eventing-natss/pkg/apis/sources/v1beta1"
	"knative.dev/pkg/kmp"
)

func TestMakeReceiveAdapter(t *testing.T) {
	src := &v1beta1.NatssSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "source-name",
			Namespace: "source-namespace",
		},
		Spec: v1beta1.NatssSourceSpec{
			Subjects:      []string{"topic1,topic2"},
			ConsumerGroup: "group",
			NatssServerSpec: sourcessv1beta1.NatssServerSpec{
				NatssURL: "SERVER1",
			},
		},
	}

	got := MakeReceiveAdapter(&ReceiveAdapterArgs{
		Image:  "test-image",
		Source: src,
		Labels: map[string]string{
			"test-key1": "test-value1",
			"test-key2": "test-value2",
		},
		SinkURI: "sink-uri",
	})

	one := int32(1)
	want := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "source-namespace",
			GenerateName: "source-name-",
			Labels: map[string]string{
				"test-key1": "test-value1",
				"test-key2": "test-value2",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"test-key1": "test-value1",
					"test-key2": "test-value2",
				},
			},
			Replicas: &one,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
					Labels: map[string]string{
						"test-key1": "test-value1",
						"test-key2": "test-value2",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "source-svc-acct",
					Containers: []corev1.Container{
						{
							Name:  "receive-adapter",
							Image: "test-image",
							Env: []corev1.EnvVar{
								{
									Name:  "NATSS_BOOTSTRAP_SERVERS",
									Value: "server1,server2",
								},
								{
									Name:  "NATSS_TOPICS",
									Value: "topic1,topic2",
								},
								{
									Name:  "NATSS_CONSUMER_GROUP",
									Value: "group",
								},
								{
									Name:  "NATSS_NET_SASL_ENABLE",
									Value: "true",
								},
								{
									Name:  "NATSS_NET_TLS_ENABLE",
									Value: "true",
								},
								{
									Name:  "SINK_URI",
									Value: "sink-uri",
								},
								{
									Name:  "NAME",
									Value: "source-name",
								},
								{
									Name:  "NAMESPACE",
									Value: "source-namespace",
								},
								{
									Name: "NATSS_NET_SASL_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-user-secret",
											},
											Key: "user",
										},
									},
								},
								{
									Name: "NATSS_NET_SASL_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-password-secret",
											},
											Key: "password",
										},
									},
								},
								{
									Name: "NATSS_NET_SASL_TYPE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-saslType-secret",
											},
											Key: "saslType",
										},
									},
								},
								{
									Name: "NATSS_NET_TLS_CERT",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-cert-secret",
											},
											Key: "tls.crt",
										},
									},
								},
								{
									Name: "NATSS_NET_TLS_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-key-secret",
											},
											Key: "tls.key",
										},
									},
								},
								{
									Name: "NATSS_NET_TLS_CA_CERT",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "the-ca-cert-secret",
											},
											Key: "tls.crt",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	if diff, err := kmp.SafeDiff(want, got); err != nil {
		t.Errorf("unexpected deploy (-want, +got) = %v", diff)
	}
}

func TestMakeReceiveAdapterNoNet(t *testing.T) {
	src := &v1beta1.NatssSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "source-name",
			Namespace: "source-namespace",
		},
		Spec: v1beta1.NatssSourceSpec{
			Subjects: []string{"topic1,topic2"},
			NatssServerSpec: sourcessv1beta1.NatssServerSpec{
				NatssURL: "server1,server2",
			},
			ConsumerGroup: "group",
		},
	}

	got := MakeReceiveAdapter(&ReceiveAdapterArgs{
		Image:  "test-image",
		Source: src,
		Labels: map[string]string{
			"test-key1": "test-value1",
			"test-key2": "test-value2",
		},
		SinkURI: "sink-uri",
	})

	one := int32(1)
	want := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "source-namespace",
			GenerateName: "source-name-",
			Labels: map[string]string{
				"test-key1": "test-value1",
				"test-key2": "test-value2",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"test-key1": "test-value1",
					"test-key2": "test-value2",
				},
			},
			Replicas: &one,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
					Labels: map[string]string{
						"test-key1": "test-value1",
						"test-key2": "test-value2",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "source-svc-acct",
					Containers: []corev1.Container{
						{
							Name:  "receive-adapter",
							Image: "test-image",
							Env: []corev1.EnvVar{
								{
									Name:  "NATSS_BOOTSTRAP_SERVERS",
									Value: "server1,server2",
								},
								{
									Name:  "NATSS_TOPICS",
									Value: "topic1,topic2",
								},
								{
									Name:  "NATSS_CONSUMER_GROUP",
									Value: "group",
								},
								{
									Name:  "NATSS_NET_SASL_ENABLE",
									Value: "false",
								},
								{
									Name:  "NATSS_NET_TLS_ENABLE",
									Value: "false",
								},
								{
									Name:  "SINK_URI",
									Value: "sink-uri",
								},
								{
									Name:  "NAME",
									Value: "source-name",
								},
								{
									Name:  "NAMESPACE",
									Value: "source-namespace",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	if diff, err := kmp.SafeDiff(want, got); err != nil {
		t.Errorf("unexpected deploy (-want, +got) = %v", diff)
	}
	want.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  "receive-adapter",
			Image: "test-image",
			Env: []corev1.EnvVar{
				{
					Name:  "NATSS_BOOTSTRAP_SERVERS",
					Value: "server1,server2",
				},
				{
					Name:  "NATSS_TOPICS",
					Value: "topic1,topic2",
				},
				{
					Name:  "NATSS_CONSUMER_GROUP",
					Value: "group",
				},
				{
					Name:  "NATSS_NET_SASL_ENABLE",
					Value: "false",
				},
				{
					Name:  "NATSS_NET_TLS_ENABLE",
					Value: "false",
				},
				{
					Name:  "SINK_URI",
					Value: "sink-uri",
				},
				{
					Name: "NATSS_NET_SASL_USER",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "the-user-secret",
							},
							Key: "user",
						},
					},
				},
				{
					Name: "NATSS_NET_SASL_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "the-password-secret",
							},
							Key: "password",
						},
					},
				},
				{
					Name: "NATSS_NET_SASL_TYPE",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "the-saslType-secret",
							},
							Key: "saslType",
						},
					},
				},
				{
					Name: "NATSS_NET_TLS_CERT",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "the-cert-secret",
							},
							Key: "tls.crt",
						},
					},
				},
				{
					Name: "NATSS_NET_TLS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "the-key-secret",
							},
							Key: "tls.key",
						},
					},
				},
				{
					Name: "NATSS_NET_TLS_CA_CERT",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "the-ca-cert-secret",
							},
							Key: "tls.crt",
						},
					},
				},
			},
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("102m"),
					corev1.ResourceMemory: resource.MustParse("500Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("101m"),
					corev1.ResourceMemory: resource.MustParse("200Mi"),
				},
			},
		},
	}

	if diff, err := kmp.SafeDiff(want, got); err != nil {
		t.Errorf("unexpected deploy (-want, +got) = %v", diff)
	}
}

func TestMakeReceiveAdapterKeyType(t *testing.T) {
	src := &v1beta1.NatssSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "source-name",
			Namespace: "source-namespace",
		},
		Spec: v1beta1.NatssSourceSpec{
			Subjects: []string{"topic1,topic2"},
			NatssServerSpec: sourcessv1beta1.NatssServerSpec{
				NatssURL: "server1,server2",
			},
			ConsumerGroup: "group",
		},
	}

	got := MakeReceiveAdapter(&ReceiveAdapterArgs{
		Image:  "test-image",
		Source: src,
		Labels: map[string]string{
			"test-key1": "test-value1",
			"test-key2": "test-value2",
		},
		SinkURI: "sink-uri",
	})

	one := int32(1)
	want := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "source-namespace",
			GenerateName: "source-name-",
			Labels: map[string]string{
				"test-key1": "test-value1",
				"test-key2": "test-value2",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"test-key1": "test-value1",
					"test-key2": "test-value2",
				},
			},
			Replicas: &one,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
					Labels: map[string]string{
						"test-key1": "test-value1",
						"test-key2": "test-value2",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "source-svc-acct",
					Containers: []corev1.Container{
						{
							Name:  "receive-adapter",
							Image: "test-image",
							Env: []corev1.EnvVar{
								{
									Name:  "NATSS_BOOTSTRAP_SERVERS",
									Value: "server1,server2",
								},
								{
									Name:  "NATSS_TOPICS",
									Value: "topic1,topic2",
								},
								{
									Name:  "NATSS_CONSUMER_GROUP",
									Value: "group",
								},
								{
									Name:  "NATSS_NET_SASL_ENABLE",
									Value: "false",
								},
								{
									Name:  "NATSS_NET_TLS_ENABLE",
									Value: "false",
								},
								{
									Name:  "SINK_URI",
									Value: "sink-uri",
								},
								{
									Name:  "NAME",
									Value: "source-name",
								},
								{
									Name:  "NAMESPACE",
									Value: "source-namespace",
								},
								{
									Name:  "KEY_TYPE",
									Value: "int",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	if diff, err := kmp.SafeDiff(want, got); err != nil {
		t.Errorf("unexpected deploy (-want, +got) = %v", diff)
	}
}
