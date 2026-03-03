/*
Copyright 2025 The KubeFleet Authors.

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

package rest

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"

	clusterv1beta1 "github.com/kubefleet-dev/kubefleet/apis/cluster/v1beta1"
)

var memberClusterPropertiesResource = schema.GroupResource{
	Group:    "clusterproperties.kubernetes-fleet.io",
	Resource: "memberclusterproperties",
}

// MemberClusterPropertiesStorage implements in-memory REST storage for
// MemberClusterProperties resources.
type MemberClusterPropertiesStorage struct {
	mu              sync.RWMutex
	data            map[string]*clusterv1beta1.MemberClusterProperties
	resourceVersion uint64
	broadcaster     *watch.Broadcaster
}

// Compile-time interface checks.
var _ rest.Storage = &MemberClusterPropertiesStorage{}
var _ rest.Scoper = &MemberClusterPropertiesStorage{}
var _ rest.Getter = &MemberClusterPropertiesStorage{}
var _ rest.Lister = &MemberClusterPropertiesStorage{}
var _ rest.Creater = &MemberClusterPropertiesStorage{}
var _ rest.Updater = &MemberClusterPropertiesStorage{}
var _ rest.GracefulDeleter = &MemberClusterPropertiesStorage{}
var _ rest.Watcher = &MemberClusterPropertiesStorage{}

// NewMemberClusterPropertiesStorage creates a new in-memory storage instance.
func NewMemberClusterPropertiesStorage() *MemberClusterPropertiesStorage {
	return &MemberClusterPropertiesStorage{
		data:        make(map[string]*clusterv1beta1.MemberClusterProperties),
		broadcaster: watch.NewBroadcaster(100, watch.WaitIfChannelFull),
	}
}

// New returns a new instance of MemberClusterProperties.
func (s *MemberClusterPropertiesStorage) New() runtime.Object {
	return &clusterv1beta1.MemberClusterProperties{}
}

// Destroy cleans up storage resources.
func (s *MemberClusterPropertiesStorage) Destroy() {
	s.broadcaster.Shutdown()
}

// NamespaceScoped returns false because MemberClusterProperties is cluster-scoped.
func (s *MemberClusterPropertiesStorage) NamespaceScoped() bool {
	return false
}

// Get retrieves a single MemberClusterProperties by name.
func (s *MemberClusterPropertiesStorage) Get(
	ctx context.Context,
	name string,
	options *metav1.GetOptions,
) (runtime.Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obj, exists := s.data[name]
	if !exists {
		return nil, errors.NewNotFound(memberClusterPropertiesResource, name)
	}
	return obj.DeepCopy(), nil
}

// NewList returns a new MemberClusterPropertiesList.
func (s *MemberClusterPropertiesStorage) NewList() runtime.Object {
	return &clusterv1beta1.MemberClusterPropertiesList{}
}

// List returns a list of MemberClusterProperties, optionally filtered by label selector.
func (s *MemberClusterPropertiesStorage) List(
	ctx context.Context,
	options *metainternalversion.ListOptions,
) (runtime.Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	labelSelector := options.LabelSelector

	list := &clusterv1beta1.MemberClusterPropertiesList{}
	for _, obj := range s.data {
		if labelSelector != nil && !labelSelector.Matches(labels.Set(obj.Labels)) {
			continue
		}
		list.Items = append(list.Items, *obj.DeepCopy())
	}

	list.ResourceVersion = strconv.FormatUint(s.resourceVersion, 10)
	return list, nil
}

// ConvertToTable implements the TableConvertor interface for kubectl output.
func (s *MemberClusterPropertiesStorage) ConvertToTable(
	ctx context.Context,
	object runtime.Object,
	tableOptions runtime.Object,
) (*metav1.Table, error) {
	return rest.NewDefaultTableConvertor(memberClusterPropertiesResource).ConvertToTable(ctx, object, tableOptions)
}

// Create stores a new MemberClusterProperties object.
func (s *MemberClusterPropertiesStorage) Create(
	ctx context.Context,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	options *metav1.CreateOptions,
) (runtime.Object, error) {
	properties, ok := obj.(*clusterv1beta1.MemberClusterProperties)
	if !ok {
		return nil, fmt.Errorf("not a MemberClusterProperties: %T", obj)
	}

	if properties.Name == "" {
		return nil, errors.NewBadRequest("name is required")
	}

	if createValidation != nil {
		if err := createValidation(ctx, obj); err != nil {
			return nil, err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[properties.Name]; exists {
		return nil, errors.NewAlreadyExists(memberClusterPropertiesResource, properties.Name)
	}

	stored := properties.DeepCopy()
	stored.CreationTimestamp = metav1.NewTime(time.Now())
	stored.ResourceVersion = s.nextResourceVersion()
	stored.Generation = 1

	s.data[stored.Name] = stored
	s.broadcaster.Action(watch.Added, stored.DeepCopy())

	return stored.DeepCopy(), nil
}

// Update modifies an existing MemberClusterProperties object.
func (s *MemberClusterPropertiesStorage) Update(
	ctx context.Context,
	name string,
	objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc,
	updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *metav1.UpdateOptions,
) (runtime.Object, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.data[name]

	if !exists {
		if !forceAllowCreate {
			return nil, false, errors.NewNotFound(memberClusterPropertiesResource, name)
		}

		// Create the object if forceAllowCreate is set.
		var oldObj runtime.Object
		updatedObj, err := objInfo.UpdatedObject(ctx, oldObj)
		if err != nil {
			return nil, false, err
		}

		if createValidation != nil {
			if err := createValidation(ctx, updatedObj); err != nil {
				return nil, false, err
			}
		}

		properties := updatedObj.(*clusterv1beta1.MemberClusterProperties)
		stored := properties.DeepCopy()
		stored.Name = name
		stored.CreationTimestamp = metav1.NewTime(time.Now())
		stored.ResourceVersion = s.nextResourceVersion()
		stored.Generation = 1

		s.data[name] = stored
		s.broadcaster.Action(watch.Added, stored.DeepCopy())

		return stored.DeepCopy(), true, nil
	}

	oldObj := existing.DeepCopy()
	updatedObj, err := objInfo.UpdatedObject(ctx, oldObj)
	if err != nil {
		return nil, false, err
	}

	properties := updatedObj.(*clusterv1beta1.MemberClusterProperties)

	// Check for optimistic concurrency via resourceVersion.
	if properties.ResourceVersion != "" && properties.ResourceVersion != existing.ResourceVersion {
		return nil, false, errors.NewConflict(
			memberClusterPropertiesResource,
			name,
			fmt.Errorf("the object has been modified; please apply your changes to the latest version"),
		)
	}

	if updateValidation != nil {
		if err := updateValidation(ctx, updatedObj, oldObj); err != nil {
			return nil, false, err
		}
	}

	stored := properties.DeepCopy()
	stored.ResourceVersion = s.nextResourceVersion()
	stored.Generation = existing.Generation + 1
	// Preserve creation timestamp from the original object.
	stored.CreationTimestamp = existing.CreationTimestamp

	s.data[name] = stored
	s.broadcaster.Action(watch.Modified, stored.DeepCopy())

	return stored.DeepCopy(), false, nil
}

// Delete removes a MemberClusterProperties object by name.
func (s *MemberClusterPropertiesStorage) Delete(
	ctx context.Context,
	name string,
	deleteValidation rest.ValidateObjectFunc,
	options *metav1.DeleteOptions,
) (runtime.Object, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.data[name]
	if !exists {
		return nil, false, errors.NewNotFound(memberClusterPropertiesResource, name)
	}

	if deleteValidation != nil {
		if err := deleteValidation(ctx, existing); err != nil {
			return nil, false, err
		}
	}

	deleted := existing.DeepCopy()
	delete(s.data, name)
	s.broadcaster.Action(watch.Deleted, deleted)

	return deleted, true, nil
}

// Watch returns a watch.Interface that watches for changes to MemberClusterProperties.
func (s *MemberClusterPropertiesStorage) Watch(
	ctx context.Context,
	options *metainternalversion.ListOptions,
) (watch.Interface, error) {
	return s.broadcaster.Watch()
}

// nextResourceVersion increments and returns the next resource version string.
// Must be called while holding the write lock.
func (s *MemberClusterPropertiesStorage) nextResourceVersion() string {
	s.resourceVersion++
	return strconv.FormatUint(s.resourceVersion, 10)
}
