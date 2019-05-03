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

package sk8

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"vmware.io/sk8/pkg/cluster"
)

type flagVals struct {
	config string
	secure bool
}

// NewCommand returns a new cobra.Command for printing a parsed configuration
func NewCommand() *cobra.Command {
	flags := flagVals{}
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(1),
		Use:   "sk8 [cluster-name]",
		Short: "Inspect the config data",
		Long:  "Inspect the config data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, cmd, args)
		},
	}

	cmd.Flags().StringVar(&flags.config, "config", "", "path to a sk8 config file")
	cmd.Flags().BoolVar(&flags.secure, "secure", true, "set to false to print credential information")
	return cmd
}

func runE(flags flagVals, cmd *cobra.Command, args []string) error {
	var name string
	if flags.config == "" {
		// If there is no cluster ID on the command line then check to
		// see if there is a default cluster ID.
		if len(args) == 0 {
			name = cluster.DefaultName()
		} else {
			name = args[0]
			if err := cluster.ValidateName(name); err != nil {
				return err
			}
		}
		if name == "" {
			return errors.Wrapf(cluster.ErrNotFound, "cluster name=%q", name)
		}
		flags.config = cluster.FilePath(name, "sk8.conf")
	}

	clu, err := cluster.ReadFromFile(flags.config)
	if err != nil {
		return errors.Wrapf(err, "config file=%q", flags.config)
	}

	if _, err := clu.WriteTo(os.Stdout); err != nil {
		return err
	}

	return nil
}
