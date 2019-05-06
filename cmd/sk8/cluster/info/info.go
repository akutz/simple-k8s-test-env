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

package info

import (
	"os"

	"github.com/spf13/cobra"

	"vmware.io/sk8/pkg/cluster"
)

type flagVals struct {
	format string
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	flags := flagVals{}
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(1),
		Use:   "info [cluster-name]",
		Short: "Print information about a cluster",
		Long:  "Print information about a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, cmd, args)
		},
	}

	cmd.Flags().StringVarP(
		&flags.format, "format", "o", "text",
		"specify the output format of the summary: json, yaml, text")

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
	clu, err := cluster.Load(args[0])
	if err != nil {
		if err == cluster.ErrNotFound {
			return nil
		}
		return err
	}
	if _, err := cluster.WithStdDefaults(clu); err != nil {
		return err
	}
	return cluster.PrintInfo(os.Stdout, flags.format, clu)
}