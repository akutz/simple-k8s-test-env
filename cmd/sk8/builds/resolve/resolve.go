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

package resolve

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vmware.io/sk8/pkg/builds"
)

type flagVals struct {
}

// NewCommand returns a new cobra.Command for printing a parsed configuration
func NewCommand() *cobra.Command {
	flags := flagVals{}
	cmd := &cobra.Command{
		Args:  cobra.MinimumNArgs(1),
		Use:   "resolve build-id [build-id...]",
		Short: "Resolves a Kubernetes build ID",
		Long:  "Resolves a Kubernetes build ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, cmd, args)
		},
	}
	return cmd
}

func runE(flags flagVals, cmd *cobra.Command, args []string) error {
	failed := false
	for _, buildID := range args {
		uri, _, err := builds.Resolve(buildID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			failed = true
		} else {
			fmt.Println(uri)
		}
	}
	if failed {
		os.Exit(1)
	}
	return nil
}
