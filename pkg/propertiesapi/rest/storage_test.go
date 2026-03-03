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
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"

	clusterv1beta1 "github.com/kubefleet-dev/kubefleet/apis/cluster/v1beta1"
)

func newTestStorage() *MemberClusterPropertiesStorage {
	return NewMemberClusterPropertiesStorage()
}

func newTestProperties(name string) *clusterv1beta1.MemberClusterProperties {
	return &clusterv1beta1.MemberClusterProperties{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newTestPropertiesWithLabels(name string, lbls map[string]string) *clusterv1beta1.MemberClusterProperties {
	return &clusterv1beta1.MemberClusterProperties{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: lbls,
		},
	}
}

func newTestPropertiesWithData(name string) *clusterv1beta1.MemberClusterProperties {
	return &clusterv1beta1.MemberClusterProperties{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Properties: map[clusterv1beta1.PropertyName]clusterv1beta1.PropertyValue{
			"kubernetes-fleet.io/node-count": {
				Value:           "3",
				ObservationTime: metav1.Now(),
			},
		},
		ResourceUsage: clusterv1beta1.ResourceUsage{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("8"),
				corev1.ResourceMemory: resource.MustParse("32Gi"),
			},
		},
	}
}

// createObj is a helper that creates an object in storage and returns it.
func createObj(t *testing.T, s *MemberClusterPropertiesStorage, obj *clusterv1beta1.MemberClusterProperties) *clusterv1beta1.MemberClusterProperties {
	t.Helper()
	result, err := s.Create(context.Background(), obj, nil, &metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Create(%q) unexpected error: %v", obj.Name, err)
	}
	return result.(*clusterv1beta1.MemberClusterProperties)
}

func TestNew(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	obj := s.New()
	got, ok := obj.(*clusterv1beta1.MemberClusterProperties)
	if !ok {
		t.Fatalf("New() = %T, want *MemberClusterProperties", obj)
	}
	want := &clusterv1beta1.MemberClusterProperties{}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("New() mismatch (-want +got):\n%s", diff)
	}
}

func TestNamespaceScoped(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	got := s.NamespaceScoped()
	if got != false {
		t.Errorf("NamespaceScoped() = %v, want false", got)
	}
}

func TestCreate(t *testing.T) {
	tests := map[string]struct {
		obj         *clusterv1beta1.MemberClusterProperties
		setup       func(*MemberClusterPropertiesStorage)
		wantErr     bool
		wantErrType func(error) bool
	}{
		"create new object succeeds": {
			obj: newTestProperties("cluster-1"),
		},
		"create with data succeeds": {
			obj: newTestPropertiesWithData("cluster-2"),
		},
		"create already existing object fails": {
			obj: newTestProperties("cluster-1"),
			setup: func(s *MemberClusterPropertiesStorage) {
				createObj(t, s, newTestProperties("cluster-1"))
			},
			wantErr:     true,
			wantErrType: errors.IsAlreadyExists,
		},
		"create with empty name fails": {
			obj:         newTestProperties(""),
			wantErr:     true,
			wantErrType: errors.IsBadRequest,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := newTestStorage()
			defer s.Destroy()

			if tc.setup != nil {
				tc.setup(s)
			}

			got, err := s.Create(context.Background(), tc.obj, nil, &metav1.CreateOptions{})
			if tc.wantErr {
				if err == nil {
					t.Errorf("Create(%q) = %v, want error", tc.obj.Name, got)
				}
				if tc.wantErrType != nil && !tc.wantErrType(err) {
					t.Errorf("Create(%q) error = %v, want matching error type", tc.obj.Name, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Create(%q) unexpected error: %v", tc.obj.Name, err)
			}
			result := got.(*clusterv1beta1.MemberClusterProperties)
			if result.Name != tc.obj.Name {
				t.Errorf("Create(%q).Name = %q, want %q", tc.obj.Name, result.Name, tc.obj.Name)
			}
			if result.ResourceVersion == "" {
				t.Errorf("Create(%q).ResourceVersion = %q, want non-empty", tc.obj.Name, result.ResourceVersion)
			}
			if result.CreationTimestamp.IsZero() {
				t.Errorf("Create(%q).CreationTimestamp is zero, want non-zero", tc.obj.Name)
			}
		})
	}
}

func TestCreateValidation(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	validationErr := errors.NewBadRequest("validation failed")
	validator := func(ctx context.Context, obj runtime.Object) error {
		return validationErr
	}

	_, err := s.Create(context.Background(), newTestProperties("cluster-1"), validator, &metav1.CreateOptions{})
	if err == nil {
		t.Fatal("Create with failing validation = nil, want error")
	}
	if err != validationErr {
		t.Errorf("Create with failing validation error = %v, want %v", err, validationErr)
	}
}

func TestGet(t *testing.T) {
	tests := map[string]struct {
		setup   func(*MemberClusterPropertiesStorage)
		name    string
		wantErr bool
	}{
		"get existing object succeeds": {
			setup: func(s *MemberClusterPropertiesStorage) {
				createObj(t, s, newTestProperties("cluster-1"))
			},
			name: "cluster-1",
		},
		"get nonexistent object fails": {
			setup:   func(s *MemberClusterPropertiesStorage) {},
			name:    "nonexistent",
			wantErr: true,
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			s := newTestStorage()
			defer s.Destroy()
			tc.setup(s)

			got, err := s.Get(context.Background(), tc.name, &metav1.GetOptions{})
			if tc.wantErr {
				if err == nil {
					t.Errorf("Get(%q) = %v, want error", tc.name, got)
				}
				if !errors.IsNotFound(err) {
					t.Errorf("Get(%q) error = %v, want NotFound error", tc.name, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get(%q) unexpected error: %v", tc.name, err)
			}
			result := got.(*clusterv1beta1.MemberClusterProperties)
			if result.Name != tc.name {
				t.Errorf("Get(%q).Name = %q, want %q", tc.name, result.Name, tc.name)
			}
		})
	}
}

func TestGetReturnsCopy(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	createObj(t, s, newTestPropertiesWithData("cluster-1"))

	got1, _ := s.Get(context.Background(), "cluster-1", &metav1.GetOptions{})
	got2, _ := s.Get(context.Background(), "cluster-1", &metav1.GetOptions{})

	p1 := got1.(*clusterv1beta1.MemberClusterProperties)
	p2 := got2.(*clusterv1beta1.MemberClusterProperties)

	// Mutating one should not affect the other.
	p1.Properties["modified"] = clusterv1beta1.PropertyValue{Value: "modified"}

	if _, exists := p2.Properties["modified"]; exists {
		t.Error("Get() should return deep copies, but mutation of one affected the other")
	}
}

func TestList(t *testing.T) {
	tests := map[string]struct {
		setup         func(*MemberClusterPropertiesStorage)
		labelSelector string
		wantCount     int
	}{
		"list all objects": {
			setup: func(s *MemberClusterPropertiesStorage) {
				createObj(t, s, newTestProperties("cluster-1"))
				createObj(t, s, newTestProperties("cluster-2"))
				createObj(t, s, newTestProperties("cluster-3"))
			},
			wantCount: 3,
		},
		"list with matching label selector": {
			setup: func(s *MemberClusterPropertiesStorage) {
				createObj(t, s, newTestPropertiesWithLabels("cluster-1", map[string]string{"env": "prod"}))
				createObj(t, s, newTestPropertiesWithLabels("cluster-2", map[string]string{"env": "staging"}))
				createObj(t, s, newTestPropertiesWithLabels("cluster-3", map[string]string{"env": "prod"}))
			},
			labelSelector: "env=prod",
			wantCount:     2,
		},
		"list with non-matching label selector": {
			setup: func(s *MemberClusterPropertiesStorage) {
				createObj(t, s, newTestPropertiesWithLabels("cluster-1", map[string]string{"env": "prod"}))
			},
			labelSelector: "env=dev",
			wantCount:     0,
		},
		"list empty storage": {
			setup:     func(s *MemberClusterPropertiesStorage) {},
			wantCount: 0,
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			s := newTestStorage()
			defer s.Destroy()
			tc.setup(s)

			opts := &metainternalversion.ListOptions{}
			if tc.labelSelector != "" {
				parsed, err := labels.Parse(tc.labelSelector)
				if err != nil {
					t.Fatalf("labels.Parse(%q) unexpected error: %v", tc.labelSelector, err)
				}
				opts.LabelSelector = parsed
			}

			got, err := s.List(context.Background(), opts)
			if err != nil {
				t.Fatalf("List() unexpected error: %v", err)
			}
			list := got.(*clusterv1beta1.MemberClusterPropertiesList)
			if len(list.Items) != tc.wantCount {
				t.Errorf("List() returned %d items, want %d", len(list.Items), tc.wantCount)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := map[string]struct {
		setup            func(*MemberClusterPropertiesStorage) *clusterv1beta1.MemberClusterProperties
		name             string
		updater          func(existing runtime.Object) runtime.Object
		forceAllowCreate bool
		wantErr          bool
		wantErrType      func(error) bool
		wantCreated      bool
	}{
		"update existing object succeeds": {
			setup: func(s *MemberClusterPropertiesStorage) *clusterv1beta1.MemberClusterProperties {
				return createObj(t, s, newTestProperties("cluster-1"))
			},
			name: "cluster-1",
			updater: func(existing runtime.Object) runtime.Object {
				updated := existing.(*clusterv1beta1.MemberClusterProperties).DeepCopy()
				updated.Properties = map[clusterv1beta1.PropertyName]clusterv1beta1.PropertyValue{
					"new-prop": {Value: "value1", ObservationTime: metav1.Now()},
				}
				return updated
			},
		},
		"update with stale resourceVersion fails": {
			setup: func(s *MemberClusterPropertiesStorage) *clusterv1beta1.MemberClusterProperties {
				return createObj(t, s, newTestProperties("cluster-1"))
			},
			name: "cluster-1",
			updater: func(existing runtime.Object) runtime.Object {
				updated := existing.(*clusterv1beta1.MemberClusterProperties).DeepCopy()
				updated.ResourceVersion = "999"
				return updated
			},
			wantErr:     true,
			wantErrType: errors.IsConflict,
		},
		"update nonexistent object without forceAllowCreate fails": {
			setup: func(s *MemberClusterPropertiesStorage) *clusterv1beta1.MemberClusterProperties {
				return nil
			},
			name: "nonexistent",
			updater: func(existing runtime.Object) runtime.Object {
				return newTestProperties("nonexistent")
			},
			wantErr:     true,
			wantErrType: errors.IsNotFound,
		},
		"update nonexistent object with forceAllowCreate creates it": {
			setup: func(s *MemberClusterPropertiesStorage) *clusterv1beta1.MemberClusterProperties {
				return nil
			},
			name: "new-cluster",
			updater: func(existing runtime.Object) runtime.Object {
				return newTestProperties("new-cluster")
			},
			forceAllowCreate: true,
			wantCreated:      true,
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			s := newTestStorage()
			defer s.Destroy()
			created := tc.setup(s)

			objInfo := &simpleUpdateObjectInfo{
				updater: func(ctx context.Context, oldObj runtime.Object) (runtime.Object, error) {
					return tc.updater(oldObj), nil
				},
			}

			got, wasCreated, err := s.Update(
				context.Background(), tc.name, objInfo,
				nil, nil,
				tc.forceAllowCreate, &metav1.UpdateOptions{},
			)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Update(%q) = %v, want error", tc.name, got)
				}
				if tc.wantErrType != nil && !tc.wantErrType(err) {
					t.Errorf("Update(%q) error = %v, want matching error type", tc.name, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Update(%q) unexpected error: %v", tc.name, err)
			}
			if wasCreated != tc.wantCreated {
				t.Errorf("Update(%q) created = %v, want %v", tc.name, wasCreated, tc.wantCreated)
			}
			result := got.(*clusterv1beta1.MemberClusterProperties)
			if result.Name != tc.name {
				t.Errorf("Update(%q).Name = %q, want %q", tc.name, result.Name, tc.name)
			}
			if !tc.wantCreated && created != nil {
				// On update, resourceVersion should increase.
				if result.ResourceVersion == created.ResourceVersion {
					t.Errorf("Update(%q).ResourceVersion = %q, want different from original %q",
						tc.name, result.ResourceVersion, created.ResourceVersion)
				}
			}
		})
	}
}

func TestUpdatePreservesCreationTimestamp(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	created := createObj(t, s, newTestProperties("cluster-1"))
	wantTimestamp := created.CreationTimestamp

	objInfo := &simpleUpdateObjectInfo{
		updater: func(ctx context.Context, oldObj runtime.Object) (runtime.Object, error) {
			updated := oldObj.(*clusterv1beta1.MemberClusterProperties).DeepCopy()
			updated.Properties = map[clusterv1beta1.PropertyName]clusterv1beta1.PropertyValue{
				"prop": {Value: "v1"},
			}
			return updated, nil
		},
	}

	got, _, err := s.Update(context.Background(), "cluster-1", objInfo, nil, nil, false, &metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}

	result := got.(*clusterv1beta1.MemberClusterProperties)
	if diff := cmp.Diff(wantTimestamp, result.CreationTimestamp, cmpopts.EquateApproxTime(time.Second)); diff != "" {
		t.Errorf("Update() CreationTimestamp mismatch (-want +got):\n%s", diff)
	}
}

func TestDelete(t *testing.T) {
	tests := map[string]struct {
		setup   func(*MemberClusterPropertiesStorage)
		name    string
		wantErr bool
	}{
		"delete existing object succeeds": {
			setup: func(s *MemberClusterPropertiesStorage) {
				createObj(t, s, newTestProperties("cluster-1"))
			},
			name: "cluster-1",
		},
		"delete nonexistent object fails": {
			setup:   func(s *MemberClusterPropertiesStorage) {},
			name:    "nonexistent",
			wantErr: true,
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			s := newTestStorage()
			defer s.Destroy()
			tc.setup(s)

			got, _, err := s.Delete(context.Background(), tc.name, nil, &metav1.DeleteOptions{})
			if tc.wantErr {
				if err == nil {
					t.Errorf("Delete(%q) = %v, want error", tc.name, got)
				}
				if !errors.IsNotFound(err) {
					t.Errorf("Delete(%q) error = %v, want NotFound error", tc.name, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Delete(%q) unexpected error: %v", tc.name, err)
			}

			// Verify the object is actually removed.
			_, getErr := s.Get(context.Background(), tc.name, &metav1.GetOptions{})
			if !errors.IsNotFound(getErr) {
				t.Errorf("Get(%q) after Delete = %v, want NotFound error", tc.name, getErr)
			}
		})
	}
}

func TestDeleteValidation(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	createObj(t, s, newTestProperties("cluster-1"))

	validationErr := errors.NewBadRequest("cannot delete")
	validator := func(ctx context.Context, obj runtime.Object) error {
		return validationErr
	}

	_, _, err := s.Delete(context.Background(), "cluster-1", validator, &metav1.DeleteOptions{})
	if err == nil {
		t.Fatal("Delete with failing validation = nil, want error")
	}
	if err != validationErr {
		t.Errorf("Delete with failing validation error = %v, want %v", err, validationErr)
	}

	// Verify the object was NOT removed.
	_, getErr := s.Get(context.Background(), "cluster-1", &metav1.GetOptions{})
	if getErr != nil {
		t.Errorf("Get after failed Delete = %v, want nil (object should still exist)", getErr)
	}
}

func TestWatchCreate(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	w, err := s.Watch(context.Background(), &metainternalversion.ListOptions{})
	if err != nil {
		t.Fatalf("Watch() unexpected error: %v", err)
	}
	defer w.Stop()

	createObj(t, s, newTestProperties("cluster-1"))

	select {
	case event := <-w.ResultChan():
		if event.Type != watch.Added {
			t.Errorf("Watch event type = %v, want %v", event.Type, watch.Added)
		}
		obj := event.Object.(*clusterv1beta1.MemberClusterProperties)
		if obj.Name != "cluster-1" {
			t.Errorf("Watch event object name = %q, want %q", obj.Name, "cluster-1")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Watch timed out waiting for Added event")
	}
}

func TestWatchUpdate(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	createObj(t, s, newTestProperties("cluster-1"))

	w, err := s.Watch(context.Background(), &metainternalversion.ListOptions{})
	if err != nil {
		t.Fatalf("Watch() unexpected error: %v", err)
	}
	defer w.Stop()

	objInfo := &simpleUpdateObjectInfo{
		updater: func(ctx context.Context, oldObj runtime.Object) (runtime.Object, error) {
			updated := oldObj.(*clusterv1beta1.MemberClusterProperties).DeepCopy()
			updated.Properties = map[clusterv1beta1.PropertyName]clusterv1beta1.PropertyValue{
				"updated-prop": {Value: "new-value"},
			}
			return updated, nil
		},
	}
	_, _, err = s.Update(context.Background(), "cluster-1", objInfo, nil, nil, false, &metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}

	select {
	case event := <-w.ResultChan():
		if event.Type != watch.Modified {
			t.Errorf("Watch event type = %v, want %v", event.Type, watch.Modified)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Watch timed out waiting for Modified event")
	}
}

func TestWatchDelete(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	createObj(t, s, newTestProperties("cluster-1"))

	w, err := s.Watch(context.Background(), &metainternalversion.ListOptions{})
	if err != nil {
		t.Fatalf("Watch() unexpected error: %v", err)
	}
	defer w.Stop()

	_, _, err = s.Delete(context.Background(), "cluster-1", nil, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}

	select {
	case event := <-w.ResultChan():
		if event.Type != watch.Deleted {
			t.Errorf("Watch event type = %v, want %v", event.Type, watch.Deleted)
		}
		obj := event.Object.(*clusterv1beta1.MemberClusterProperties)
		if obj.Name != "cluster-1" {
			t.Errorf("Watch event object name = %q, want %q", obj.Name, "cluster-1")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Watch timed out waiting for Deleted event")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("cluster-%d", id)

			// Create.
			obj := newTestPropertiesWithData(name)
			_, err := s.Create(context.Background(), obj, nil, &metav1.CreateOptions{})
			if err != nil {
				return
			}

			// Get.
			_, _ = s.Get(context.Background(), name, &metav1.GetOptions{})

			// List.
			_, _ = s.List(context.Background(), &metainternalversion.ListOptions{})

			// Update.
			objInfo := &simpleUpdateObjectInfo{
				updater: func(ctx context.Context, oldObj runtime.Object) (runtime.Object, error) {
					if oldObj == nil {
						return newTestProperties(name), nil
					}
					updated := oldObj.(*clusterv1beta1.MemberClusterProperties).DeepCopy()
					updated.Properties = map[clusterv1beta1.PropertyName]clusterv1beta1.PropertyValue{
						"concurrent": {Value: "test"},
					}
					return updated, nil
				},
			}
			_, _, _ = s.Update(context.Background(), name, objInfo, nil, nil, false, &metav1.UpdateOptions{})

			// Delete.
			_, _, _ = s.Delete(context.Background(), name, nil, &metav1.DeleteOptions{})
		}(i)
	}

	wg.Wait()
}

func TestConvertToTable(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	created := createObj(t, s, newTestProperties("cluster-1"))

	table, err := s.ConvertToTable(context.Background(), created, nil)
	if err != nil {
		t.Fatalf("ConvertToTable() unexpected error: %v", err)
	}
	if table == nil {
		t.Fatal("ConvertToTable() = nil, want non-nil table")
	}
}

func TestNewList(t *testing.T) {
	s := newTestStorage()
	defer s.Destroy()

	obj := s.NewList()
	_, ok := obj.(*clusterv1beta1.MemberClusterPropertiesList)
	if !ok {
		t.Fatalf("NewList() = %T, want *MemberClusterPropertiesList", obj)
	}
}

// simpleUpdateObjectInfo implements rest.UpdatedObjectInfo for testing.
type simpleUpdateObjectInfo struct {
	updater func(ctx context.Context, oldObj runtime.Object) (runtime.Object, error)
}

var _ rest.UpdatedObjectInfo = &simpleUpdateObjectInfo{}

func (i *simpleUpdateObjectInfo) UpdatedObject(ctx context.Context, oldObj runtime.Object) (runtime.Object, error) {
	return i.updater(ctx, oldObj)
}

func (i *simpleUpdateObjectInfo) Preconditions() *metav1.Preconditions {
	return nil
}
