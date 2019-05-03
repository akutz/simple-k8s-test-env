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
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LinuxVirtualSwitchConfig describes how to access machines via a Linux
// Virtual Switch (LVS) host.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LinuxVirtualSwitchConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`

	// PublicNIC is the name of the public network interface device.
	//
	// Defaults to "eth0"
	// +optional
	PublicNIC string `json:"publicNIC,omitempty"`

	// PublicIPv4Addr is the public IP address and is also used as the
	// virtual IP for services created on this LVS host.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	PublicIPAddr net.IP `json:"publicIPAddr,omitempty"`

	// PrivateIPv4Addr is the IP address the LVS host uses to communicate
	// with the nodes. This address also acts as the default gateway for
	// the nodes' private interfaces.
	//
	// When LVS is configured, the value of this field is automatically
	// assigned to Network.Interfaces[1].GatewayIPv4Addr. Nodes should
	// connect their second network interface as the one that receives
	// traffic from the LVS director.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	PrivateIPAddr net.IP `json:"privateIPAddr,omitempty"`

	// SSH is information used to access the LVS server.
	SSH SSHCredentialAndEndpoint `json:"ssh"`
}

// SetDefaults_LinuxVirtualSwitchConfig sets uninitialized fields to their default value.
func SetDefaults_LinuxVirtualSwitchConfig(obj *LinuxVirtualSwitchConfig) {
	if obj.PublicNIC == "" {
		obj.PublicNIC = "eth0"
	}
}
