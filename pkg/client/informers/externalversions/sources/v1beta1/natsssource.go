/*
Copyright 2021 The Knative Authors

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
	"context"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	sourcesv1beta1 "knative.dev/eventing-natss/pkg/apis/sources/v1beta1"
	versioned "knative.dev/eventing-natss/pkg/client/clientset/versioned"
	internalinterfaces "knative.dev/eventing-natss/pkg/client/informers/externalversions/internalinterfaces"
	v1beta1 "knative.dev/eventing-natss/pkg/client/listers/sources/v1beta1"
)

// NatssSourceInformer provides access to a shared informer and lister for
// NatssSources.
type NatssSourceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.NatssSourceLister
}

type natssSourceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewNatssSourceInformer constructs a new informer for NatssSource type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNatssSourceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNatssSourceInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredNatssSourceInformer constructs a new informer for NatssSource type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNatssSourceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SourcesV1beta1().NatssSources(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SourcesV1beta1().NatssSources(namespace).Watch(context.TODO(), options)
			},
		},
		&sourcesv1beta1.NatssSource{},
		resyncPeriod,
		indexers,
	)
}

func (f *natssSourceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNatssSourceInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *natssSourceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&sourcesv1beta1.NatssSource{}, f.defaultInformer)
}

func (f *natssSourceInformer) Lister() v1beta1.NatssSourceLister {
	return v1beta1.NewNatssSourceLister(f.Informer().GetIndexer())
}
