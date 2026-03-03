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
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	propertiesrest "github.com/kubefleet-dev/kubefleet/pkg/propertiesapi/rest"
)

// Config holds configuration for the properties API server.
type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
}

// PropertiesAPIServer contains state for the properties API server.
type PropertiesAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

// CompletedConfig wraps a completed server configuration.
type CompletedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
}

// Complete fills in any fields not set that are required to have valid data.
func (c *Config) Complete() CompletedConfig {
	return CompletedConfig{
		GenericConfig: c.GenericConfig.Complete(),
	}
}

// New creates a new PropertiesAPIServer from the completed configuration.
func (c CompletedConfig) New() (*PropertiesAPIServer, error) {
	genericServer, err := c.GenericConfig.New(
		"properties-api-server",
		genericapiserver.NewEmptyDelegate(),
	)
	if err != nil {
		return nil, err
	}

	s := &PropertiesAPIServer{
		GenericAPIServer: genericServer,
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(
		SchemeGroupVersion.Group,
		Scheme,
		ParameterCodec,
		Codecs,
	)

	storage := propertiesrest.NewMemberClusterPropertiesStorage()

	apiGroupInfo.VersionedResourcesStorageMap[SchemeGroupVersion.Version] = map[string]rest.Storage{
		"memberclusterproperties": storage,
	}

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
