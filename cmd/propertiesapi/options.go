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

package main

import (
	"github.com/spf13/pflag"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"

	"github.com/kubefleet-dev/kubefleet/pkg/propertiesapi"
)

// PropertiesAPIServerOptions holds options for the properties API server.
type PropertiesAPIServerOptions struct {
	RecommendedOptions *options.RecommendedOptions
}

// NewPropertiesAPIServerOptions creates default options.
func NewPropertiesAPIServerOptions() *PropertiesAPIServerOptions {
	return &PropertiesAPIServerOptions{
		RecommendedOptions: options.NewRecommendedOptions(
			"",
			propertiesapi.Codecs.LegacyCodec(propertiesapi.SchemeGroupVersion),
		),
	}
}

// AddFlags adds flags for the server options to the provided FlagSet.
func (o *PropertiesAPIServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)
}

// Config returns a completed properties API server config from the options.
func (o *PropertiesAPIServerOptions) Config() (*propertiesapi.Config, error) {
	// Disable etcd since we use in-memory storage.
	o.RecommendedOptions.Etcd = nil

	serverConfig := genericapiserver.NewRecommendedConfig(propertiesapi.Codecs)

	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	return &propertiesapi.Config{
		GenericConfig: serverConfig,
	}, nil
}
