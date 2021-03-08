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

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1beta1 "knative.dev/eventing-natss/pkg/apis/sources/v1beta1"
	scheme "knative.dev/eventing-natss/pkg/client/clientset/versioned/scheme"
)

// NatssSourcesGetter has a method to return a NatssSourceInterface.
// A group's client should implement this interface.
type NatssSourcesGetter interface {
	NatssSources(namespace string) NatssSourceInterface
}

// NatssSourceInterface has methods to work with NatssSource resources.
type NatssSourceInterface interface {
	Create(ctx context.Context, natssSource *v1beta1.NatssSource, opts v1.CreateOptions) (*v1beta1.NatssSource, error)
	Update(ctx context.Context, natssSource *v1beta1.NatssSource, opts v1.UpdateOptions) (*v1beta1.NatssSource, error)
	UpdateStatus(ctx context.Context, natssSource *v1beta1.NatssSource, opts v1.UpdateOptions) (*v1beta1.NatssSource, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.NatssSource, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.NatssSourceList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.NatssSource, err error)
	GetScale(ctx context.Context, natssSourceName string, options v1.GetOptions) (*autoscalingv1.Scale, error)

	NatssSourceExpansion
}

// natssSources implements NatssSourceInterface
type natssSources struct {
	client rest.Interface
	ns     string
}

// newNatssSources returns a NatssSources
func newNatssSources(c *SourcesV1beta1Client, namespace string) *natssSources {
	return &natssSources{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the natssSource, and returns the corresponding natssSource object, and an error if there is any.
func (c *natssSources) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.NatssSource, err error) {
	result = &v1beta1.NatssSource{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("natsssources").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NatssSources that match those selectors.
func (c *natssSources) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.NatssSourceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.NatssSourceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("natsssources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested natssSources.
func (c *natssSources) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("natsssources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a natssSource and creates it.  Returns the server's representation of the natssSource, and an error, if there is any.
func (c *natssSources) Create(ctx context.Context, natssSource *v1beta1.NatssSource, opts v1.CreateOptions) (result *v1beta1.NatssSource, err error) {
	result = &v1beta1.NatssSource{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("natsssources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(natssSource).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a natssSource and updates it. Returns the server's representation of the natssSource, and an error, if there is any.
func (c *natssSources) Update(ctx context.Context, natssSource *v1beta1.NatssSource, opts v1.UpdateOptions) (result *v1beta1.NatssSource, err error) {
	result = &v1beta1.NatssSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("natsssources").
		Name(natssSource.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(natssSource).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *natssSources) UpdateStatus(ctx context.Context, natssSource *v1beta1.NatssSource, opts v1.UpdateOptions) (result *v1beta1.NatssSource, err error) {
	result = &v1beta1.NatssSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("natsssources").
		Name(natssSource.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(natssSource).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the natssSource and deletes it. Returns an error if one occurs.
func (c *natssSources) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("natsssources").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *natssSources) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("natsssources").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched natssSource.
func (c *natssSources) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.NatssSource, err error) {
	result = &v1beta1.NatssSource{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("natsssources").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// GetScale takes name of the natssSource, and returns the corresponding autoscalingv1.Scale object, and an error if there is any.
func (c *natssSources) GetScale(ctx context.Context, natssSourceName string, options v1.GetOptions) (result *autoscalingv1.Scale, err error) {
	result = &autoscalingv1.Scale{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("natsssources").
		Name(natssSourceName).
		SubResource("scale").
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}
