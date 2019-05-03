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

package util

import (
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"vmware.io/sk8/pkg/config/encoding"
	"vmware.io/sk8/pkg/provider/vsphere/config"
)

// GetClusterProviderConfig returns the vSphere ClusterProviderConfig
// object from the CAPI cluster.
func GetClusterProviderConfig(obj *capi.Cluster) *config.ClusterProviderConfig {
	cfg := obj.Spec.ProviderSpec.Value.Object.(*config.ClusterProviderConfig)
	// Decode the NAT config from the raw data into a runtime.Object.
	encoding.FromRaw(cfg.NAT)
	return cfg
}

// GetMachineProviderConfig returns the vSphere MachineProviderConfig
// object from the CAPI machine.
func GetMachineProviderConfig(obj *capi.Machine) *config.MachineProviderConfig {
	return obj.Spec.ProviderSpec.Value.Object.(*config.MachineProviderConfig)
}
