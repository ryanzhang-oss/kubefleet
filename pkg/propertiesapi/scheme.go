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

package propertiesapi

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	clusterv1beta1 "github.com/kubefleet-dev/kubefleet/apis/cluster/v1beta1"
)

var (
	// SchemeGroupVersion is the group version served by the aggregated API server.
	SchemeGroupVersion = schema.GroupVersion{
		Group:   "clusterproperties.kubernetes-fleet.io",
		Version: "v1beta1",
	}

	// Scheme is the runtime scheme for the properties API server.
	Scheme = runtime.NewScheme()

	// Codecs provides encoders and decoders for the scheme.
	Codecs = serializer.NewCodecFactory(Scheme)

	// ParameterCodec handles conversion for query parameters.
	ParameterCodec = runtime.NewParameterCodec(Scheme)
)

func init() {
	Scheme.AddKnownTypes(SchemeGroupVersion,
		&clusterv1beta1.MemberClusterProperties{},
		&clusterv1beta1.MemberClusterPropertiesList{},
	)
	metav1.AddToGroupVersion(Scheme, SchemeGroupVersion)
}
