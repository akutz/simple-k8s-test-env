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

package config

import (
	"github.com/spf13/cobra"

	"vmware.io/sk8/cmd/sk8/config/k8s"
	"vmware.io/sk8/cmd/sk8/config/sk8"
	"vmware.io/sk8/cmd/sk8/config/ssh"
)

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "config",
		Short: "Inspect config data",
		Long:  "Inspect config data",
	}
	cmd.AddCommand(k8s.NewCommand())
	cmd.AddCommand(sk8.NewCommand())
	cmd.AddCommand(ssh.NewCommand())
	return cmd
}
