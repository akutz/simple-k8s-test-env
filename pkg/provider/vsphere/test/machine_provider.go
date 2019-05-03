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

package test

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/config/encoding"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
)

// NewExpectedMachineProviderConfig returns a new MachineProviderConfig
// object used for testing.
func NewExpectedMachineProviderConfig() *vconfig.MachineProviderConfig {
	obj := vconfig.MachineProviderConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: vconfig.SchemeGroupVersion.String(),
			Kind:       "MachineProviderConfig",
		},
		Datacenter:   "/SDDC-Datacenter",
		Datastore:    "/SDDC-Datacenter/datastore/WorkloadDatastore",
		Folder:       "/SDDC-Datacenter/vm/Workloads/sk8e2e",
		ResourcePool: "/SDDC-Datacenter/host/Cluster-1/Resources/Compute-ResourcePool/sk8e2e",
		Network: config.MachineNetworkConfig{
			Interfaces: []config.NetworkInterfaceConfig{
				{
					Name:    "eth0",
					Network: "sddc-cgw-network-4",
				},
				{
					Name:    "eth1",
					Network: "sddc-cgw-network-lvs-1",
				},
			},
		},
	}
	encoding.Scheme.Default(&obj)
	return &obj
}
