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

package list

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vmware.io/sk8/pkg/cluster"
)

type flagVals struct {
	format string
}

// NewCommand returns a new cobra.Command for printing a parsed configuration
func NewCommand() *cobra.Command {
	flags := flagVals{}
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List the known clusters",
		Long:    "List the known clusters",
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
	clusters, err := cluster.List()
	if err != nil {
		return err
	}
	if flags.format == "json" {
		fmt.Fprintf(os.Stdout, "[\n")
	}
	for i, c := range clusters {
		err := cluster.PrintInfo(os.Stdout, flags.format, c)
		if err != nil {
			return err
		}
		if i < len(clusters)-1 {
			switch flags.format {
			case "json":
				fmt.Fprintf(os.Stdout, ",\n")
			case "yaml":
				fmt.Fprintf(os.Stdout, "\n---\n\n")
			default:
				fmt.Fprintf(os.Stdout, "\n")
			}
		}
	}
	if flags.format == "json" {
		fmt.Fprintf(os.Stdout, "]\n")
	}

	return nil
}
