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

package cloudinit

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/vim25/types"
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	vutil "vmware.io/sk8/pkg/provider/vsphere/util"
)

// New returns a new cloud-init configuration for a VM.
func New(
	ctx context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine) ([]types.BaseOptionValue, error) {

	cluProCfg := vutil.GetClusterProviderConfig(cluster)
	macProCfg := vutil.GetMachineProviderConfig(machine)

	netCfg, err := GetNetworkConfig(ctx, machine, macProCfg.Network)
	if err != nil {
		return nil, errors.Wrap(
			err, "error getting cloud init network data")
	}

	metadata, err := GetMetadata(ctx, cluster, machine, netCfg)
	if err != nil {
		return nil, errors.Wrap(
			err, "error getting cloud init metadata")
	}

	userData, err := GetUserData(ctx, cluster, machine, cluProCfg.SSH)
	if err != nil {
		return nil, errors.Wrap(
			err, "error getting cloud init user data")
	}

	var exCfg extraConfig
	if err := exCfg.SetMetadata(metadata); err != nil {
		return nil, errors.Wrap(
			err, "error setting cloud init metadata")
	}
	if err := exCfg.SetUserData(userData); err != nil {
		return nil, errors.Wrap(
			err, "error setting cloud init user data")
	}

	return exCfg, nil
}
