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

package k8s

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"vmware.io/sk8/pkg/cluster"
)

type flagVals struct {
	show bool
}

// NewCommand returns a new cobra.Command for printing a parsed configuration
func NewCommand() *cobra.Command {
	flags := flagVals{}
	cmd := &cobra.Command{
		Args:    cobra.MaximumNArgs(1),
		Use:     "kube [cluster-name]",
		Aliases: []string{"k8s"},
		Short:   "Inspect the kubeconfig file",
		Long:    "Inspect the kubeconfig file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, cmd, args)
		},
	}

	cmd.Flags().BoolVar(&flags.show, "show", false, "set to true to print the file's contents")
	return cmd
}

func runE(flags flagVals, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if name := cluster.DefaultName(); name != "" {
			args = append(args, name)
		}
	}

	if len(args) == 0 {
		return nil
	}

	clusterName := args[0]
	if clusterName == "" {
		return nil
	}

	// Get the kubeconfig for the provided cluster ID.
	filePath := cluster.FilePath(clusterName, "kube.conf")
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if flags.show {
		return printToStdout(file)
	}

	fmt.Println(filePath)
	return nil
}

func printToStdout(r io.Reader) error {
	_, err := io.Copy(os.Stdout, r)
	return err
}
