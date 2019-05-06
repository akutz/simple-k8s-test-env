/*
simple-kubernetes-test-environment

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

package down

import (
	"github.com/spf13/cobra"

	"vmware.io/sk8/pkg/cluster"
	"vmware.io/sk8/pkg/cluster/delete"
	"vmware.io/sk8/pkg/status"
)

type flagVals struct {
	config string
}

// NewCommand returns a new cobra.Command for cluster deletion
func NewCommand() *cobra.Command {
	flags := flagVals{}
	cmd := &cobra.Command{
		Use:   "down [cluster-name...]",
		Short: "Turns down a Kubernetes cluster",
		Long:  "Turns down a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, cmd, args)
		},
	}

	cmd.Flags().StringVar(&flags.config, "config", "", "path to a sk8 config file")
	return cmd
}

func runE(flags flagVals, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		name := cluster.DefaultName()
		if name == "" {
			return nil
		}
		args = append(args, name)
	}

	for _, name := range args {
		clu, err := cluster.Load(name)
		if err != nil {
			if err == cluster.ErrNotFound {
				continue
			}
			return err
		}
		if err := delete.Cluster(status.Context(), clu); err != nil {
			return err
		}
		if err := clu.WriteToDisk(); err != nil {
			return err
		}
	}
	return nil
}
