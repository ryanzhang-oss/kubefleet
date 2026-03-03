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
	"flag"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/cli"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)

	opts := NewPropertiesAPIServerOptions()

	cmd := &cobra.Command{
		Use:   "properties-api-server",
		Short: "Launch the KubeFleet properties API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := opts.Config()
			if err != nil {
				return err
			}

			completed := config.Complete()
			server, err := completed.New()
			if err != nil {
				return err
			}

			ctx := genericapiserver.SetupSignalContext()
			return server.GenericAPIServer.PrepareRun().RunWithContext(ctx)
		},
	}

	opts.AddFlags(cmd.Flags())
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	code := cli.Run(cmd)
	os.Exit(code)
}
