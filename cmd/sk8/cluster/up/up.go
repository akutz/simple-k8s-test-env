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

package up

import (
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"

	"vmware.io/sk8/cmd/sk8/cluster/info"
	"vmware.io/sk8/pkg/cluster"
	"vmware.io/sk8/pkg/cluster/create"
	"vmware.io/sk8/pkg/config"
	vsphere "vmware.io/sk8/pkg/provider/vsphere/config"
	"vmware.io/sk8/pkg/status"
)

var (
	defaultRole = config.MachineRoleControlPlane | config.MachineRoleWorker
	validRoles  = config.MachineRoleControlPlane | config.MachineRoleWorker
)

type flagVals struct {
	dryRun  bool
	roles   []string
	timeout time.Duration
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	flags := flagVals{}
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(1),
		Use:   "up [build-id]",
		Short: "Turns up a Kubernetes cluster",
		Long:  "Turns up a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, cmd, args)
		},
	}

	cmd.Flags().BoolVar(
		&flags.dryRun, "dry-run", false,
		"specify this flag to emit the calculated cluster manifest without "+
			"actually applying any changes")
	cmd.Flags().DurationVar(
		&flags.timeout, "timeout", time.Minute*15,
		"amount of time to wait for the cluster to come online")
	cmd.Flags().StringArrayVarP(
		&flags.roles,
		"role",
		"r",
		[]string{defaultRole.String()},
		"the role to assign to a cluster machine. "+
			"this flag may be used multiple times to create multiple "+
			"machines. valid roles include: "+
			validRoles.String())

	return cmd
}

func runE(flags flagVals, cmd *cobra.Command, args []string) error {
	roles := make([]config.MachineRole, len(flags.roles))
	for i, r := range flags.roles {
		if err := roles[i].UnmarshalText([]byte(r)); err != nil {
			return err
		}
	}

	clu, err := cluster.NewFromRoles(
		vsphere.SchemeGroupVersion.WithKind("ClusterProviderConfig"),
		vsphere.SchemeGroupVersion.WithKind("MachineProviderConfig"),
		roles...)
	if err != nil {
		return err
	}

	if _, err := clu.WithNewName(); err != nil {
		return err
	}

	if len(args) == 0 {
		args = []string{"release/stable"}
	}
	if _, err := clu.WithKubernetesBuildID(args[0]); err != nil {
		return err
	}

	if _, err := cluster.WithStdDefaults(clu); err != nil {
		return err
	}

	if flags.dryRun {
		_, err := clu.WriteTo(os.Stdout)
		return err
	}

	confFileDir := cluster.FilePath(clu.Cluster.Name)
	os.MkdirAll(confFileDir, 0750)
	confFilePath := path.Join(confFileDir, "sk8.conf")
	if _, err := clu.WriteToFile(confFilePath); err != nil {
		return err
	}

	if err := create.Cluster(status.Context(), clu); err != nil {
		return err
	}

	return info.Print(clu)
}
