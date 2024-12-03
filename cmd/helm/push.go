/*
Copyright The Helm Authors.

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
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"helm.sh/helm/v4/cmd/helm/require"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/pusher"
)

const pushDesc = `
Upload a chart to a registry.

If the chart has an associated provenance file,
it will also be uploaded.
`

type registryPushOptions struct {
	cfg action.RegistryConfiguration
}

func newPushCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	o := &registryPushOptions{}

	cmd := &cobra.Command{
		Use:   "push [chart] [remote]",
		Short: "push a chart to remote",
		Long:  pushDesc,
		Args:  require.MinimumNArgs(2),
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				// Do file completion for the chart file to push
				return nil, cobra.ShellCompDirectiveDefault
			}
			if len(args) == 1 {
				providers := []pusher.Provider(pusher.All(settings))
				var comps []string
				for _, p := range providers {
					for _, scheme := range p.Schemes {
						comps = append(comps, fmt.Sprintf("%s://", scheme))
					}
				}
				return comps, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
			}
			return noMoreArgsComp()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			registryClient, err := o.cfg.NewClient()
			if err != nil {
				return fmt.Errorf("missing registry client: %w", err)
			}
			cfg.RegistryClient = registryClient
			chartRef := args[0]
			remote := args[1]
			client := action.NewPushWithOpts(
				action.WithPushConfig(cfg),
				action.WithTLSClientConfig(o.cfg.CertFile, o.cfg.KeyFile, o.cfg.CaFile),
				action.WithInsecureSkipTLSVerify(o.cfg.InsecureSkipTLSverify),
				action.WithPlainHTTP(o.cfg.PlainHTTP),
				action.WithPushOptWriter(out),
			)
			client.Settings = settings
			output, err := client.Run(chartRef, remote)
			if err != nil {
				return err
			}
			fmt.Fprint(out, output)
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&o.cfg.CertFile, "cert-file", "", "identify registry client using this SSL certificate file")
	f.StringVar(&o.cfg.KeyFile, "key-file", "", "identify registry client using this SSL key file")
	f.StringVar(&o.cfg.CaFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
	f.BoolVar(&o.cfg.InsecureSkipTLSverify, "insecure-skip-tls-verify", false, "skip tls certificate checks for the chart upload")
	f.BoolVar(&o.cfg.PlainHTTP, "plain-http", false, "use insecure HTTP connections for the chart upload")
	f.StringVar(&o.cfg.Username, "username", "", "chart repository username where to locate the requested chart")
	f.StringVar(&o.cfg.Password, "password", "", "chart repository password where to locate the requested chart")

	return cmd
}
