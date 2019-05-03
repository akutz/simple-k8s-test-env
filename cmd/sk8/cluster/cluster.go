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

package cluster

import (
	"github.com/spf13/cobra"

	"vmware.io/sk8/cmd/sk8/cluster/down"
	"vmware.io/sk8/cmd/sk8/cluster/info"
	"vmware.io/sk8/cmd/sk8/cluster/list"
	"vmware.io/sk8/cmd/sk8/cluster/up"
)

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "cluster",
		Short: "Manage a cluster",
		Long:  "Manage a cluster",
	}
	cmd.AddCommand(up.NewCommand())
	cmd.AddCommand(down.NewCommand())
	cmd.AddCommand(list.NewCommand())
	cmd.AddCommand(info.NewCommand())
	return cmd
}
