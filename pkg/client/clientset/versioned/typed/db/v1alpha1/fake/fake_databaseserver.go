/*
Copyright 2018 Infoblox, Inc.

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

package fake

import (
	v1alpha1 "github.com/infobloxopen/atlas-db/pkg/apis/db/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDatabaseServers implements DatabaseServerInterface
type FakeDatabaseServers struct {
	Fake *FakeAtlasdbV1alpha1
	ns   string
}

var databaseserversResource = schema.GroupVersionResource{Group: "atlasdb.infoblox.com", Version: "v1alpha1", Resource: "databaseservers"}

var databaseserversKind = schema.GroupVersionKind{Group: "atlasdb.infoblox.com", Version: "v1alpha1", Kind: "DatabaseServer"}

// Get takes name of the databaseServer, and returns the corresponding databaseServer object, and an error if there is any.
func (c *FakeDatabaseServers) Get(name string, options v1.GetOptions) (result *v1alpha1.DatabaseServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(databaseserversResource, c.ns, name), &v1alpha1.DatabaseServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DatabaseServer), err
}

// List takes label and field selectors, and returns the list of DatabaseServers that match those selectors.
func (c *FakeDatabaseServers) List(opts v1.ListOptions) (result *v1alpha1.DatabaseServerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(databaseserversResource, databaseserversKind, c.ns, opts), &v1alpha1.DatabaseServerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DatabaseServerList{}
	for _, item := range obj.(*v1alpha1.DatabaseServerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested databaseServers.
func (c *FakeDatabaseServers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(databaseserversResource, c.ns, opts))

}

// Create takes the representation of a databaseServer and creates it.  Returns the server's representation of the databaseServer, and an error, if there is any.
func (c *FakeDatabaseServers) Create(databaseServer *v1alpha1.DatabaseServer) (result *v1alpha1.DatabaseServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(databaseserversResource, c.ns, databaseServer), &v1alpha1.DatabaseServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DatabaseServer), err
}

// Update takes the representation of a databaseServer and updates it. Returns the server's representation of the databaseServer, and an error, if there is any.
func (c *FakeDatabaseServers) Update(databaseServer *v1alpha1.DatabaseServer) (result *v1alpha1.DatabaseServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(databaseserversResource, c.ns, databaseServer), &v1alpha1.DatabaseServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DatabaseServer), err
}

// Delete takes name of the databaseServer and deletes it. Returns an error if one occurs.
func (c *FakeDatabaseServers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(databaseserversResource, c.ns, name), &v1alpha1.DatabaseServer{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDatabaseServers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(databaseserversResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.DatabaseServerList{})
	return err
}

// Patch applies the patch and returns the patched databaseServer.
func (c *FakeDatabaseServers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DatabaseServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(databaseserversResource, c.ns, name, data, subresources...), &v1alpha1.DatabaseServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DatabaseServer), err
}