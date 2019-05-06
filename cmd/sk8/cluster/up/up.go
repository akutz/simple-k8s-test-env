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
	"time"

	"github.com/spf13/cobra"

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
	dryRun        bool
	roles         []string
	timeout       time.Duration
	format        string
	buildID       string
	cloudProvider string
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	flags := flagVals{}
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(1),
		Use:   "up [cluster-name]",
		Short: "Turns up a Kubernetes cluster",
		Long:  "Turns up a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, cmd, args)
		},
	}

	cmd.Flags().StringVarP(
		&flags.buildID, "build-id", "b",
		"release/stable",
		"the Kubernetes build ID to deploy: release/*, ci/*, semver")
	cmd.Flags().StringVar(
		&flags.cloudProvider, "cloud-provider",
		"",
		"the name of a cloud provider; ex. aws, vsphere, external, etc.")
	cmd.Flags().BoolVar(
		&flags.dryRun, "dry-run", false,
		"specify this flag to emit the calculated cluster manifest without "+
			"actually applying any changes")
	cmd.Flags().StringVarP(
		&flags.format, "format", "o", "text",
		"specify the output format of the summary: json, yaml, text")
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
	var clu *cluster.Cluster
	if len(args) == 1 {
		clu2, err := cluster.Load(args[0])
		if err != nil {
			return err
		}
		clu = clu2
	}
	if clu == nil {
		roles := make([]config.MachineRole, len(flags.roles))
		for i, r := range flags.roles {
			if err := roles[i].UnmarshalText([]byte(r)); err != nil {
				return err
			}
		}
		clu2, err := cluster.NewFromRoles(
			vsphere.SchemeGroupVersion.WithKind("ClusterProviderConfig"),
			vsphere.SchemeGroupVersion.WithKind("MachineProviderConfig"),
			roles...)
		if err != nil {
			return err
		}
		if _, err := clu2.WithNewName(); err != nil {
			return err
		}
		if _, err := clu2.WithKubernetesBuildID(flags.buildID); err != nil {
			return err
		}
		if _, err := cluster.WithStdDefaults(clu2); err != nil {
			return err
		}
		clu2.WithCloudProvider(flags.cloudProvider)
		clu = clu2
	}

	if flags.dryRun {
		_, err := clu.WriteTo(os.Stdout)
		return err
	}

	if err := clu.WriteToDisk(); err != nil {
		return err
	}

	if err := create.Cluster(status.Context(), clu); err != nil {
		return err
	}

	if err := clu.WriteToDisk(); err != nil {
		return err
	}

	if flags.format == "text" {
		flags.format = infoTemplate
	}

	return cluster.PrintInfo(os.Stdout, flags.format, clu)
}

const infoTemplate = cluster.DefaultTemplate + `
Print the nodes with the following command:
  kubectl --kubeconfig {{.Kubeconfig}} get nodes

Query the state of the Kubernetes system components:
  kubectl --kubeconfig {{.Kubeconfig}} -n kube-system get all

Finally, the cluster may be deleted with:
  {{.Program}} cluster down {{.Name}}
`
