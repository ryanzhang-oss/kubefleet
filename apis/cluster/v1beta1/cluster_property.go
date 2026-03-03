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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories={fleet,fleet-cluster}

// MemberClusterProperties represents the observed properties and resource usage
// of a member cluster in the fleet.
type MemberClusterProperties struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Properties is a collection of non-resource properties observed for the member cluster.
	// +optional
	Properties map[PropertyName]PropertyValue `json:"properties,omitempty"`

	// The current observed resource usage of the member cluster.
	// +optional
	ResourceUsage ResourceUsage `json:"resourceUsage,omitempty"`
}

//+kubebuilder:object:root=true

// MemberClusterPropertiesList contains a list of MemberClusterProperties.
type MemberClusterPropertiesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MemberClusterProperties `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MemberClusterProperties{}, &MemberClusterPropertiesList{})
}
