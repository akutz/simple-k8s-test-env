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
	"bytes"
	"fmt"
	"net"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachineNetworkConfig describes how to configure networking for a machine.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MachineNetworkConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`

	// Interfaces is a list of a machine's network interfaces.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Interfaces []NetworkInterfaceConfig `json:"interfaces,omitempty"`

	// Nameservers is a list of the DNS servers used by the machine. No
	// more than three nameservers should be defined.
	//
	// Defaults to 8.8.8.8, 8.8.4.4
	//
	// +optional
	Nameservers []string `json:"nameservers,omitempty"`
}

// SetDefaults_MachineNetworkConfig sets uninitialized fields to their default value.
func SetDefaults_MachineNetworkConfig(obj *MachineNetworkConfig) {
	if len(obj.Nameservers) == 0 {
		obj.Nameservers = []string{"8.8.8.8", "8.8.4.4"}
	}
}

// NetworkInterfaceConfig describes how to configure a network interface.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NetworkInterfaceConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`

	// Name is the name of the network interface, ex. eth0.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Name string `json:"name,omitempty"`

	// Network is the name of the network to which this interface should be
	// attached. This field is MachineProvider dependent.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Network string `json:"network,omitempty"`

	// Type describes whether an interface uses static addresses or DHCP.
	//
	// Defaults to "dhcp"
	//
	// +optional
	Type string `json:"type,omitempty"`

	// IPv4Addr is the IP address configured for the interface. This is
	// only used for Type=static.
	//
	// +optional
	IPAddr net.IP `json:"ipAddr,omitempty"`

	// GatewayIPv4Addr is the IP address of the gateway configured for
	// the interface. When using DHCP it is generally okay to ignore
	// this field.
	//
	// +optional
	GatewayIPAddr net.IP `json:"gatewayIPAddr,omitempty"`
}

// SetDefaults_NetworkInterfaceConfig sets uninitialized fields to their default value.
func SetDefaults_NetworkInterfaceConfig(obj *NetworkInterfaceConfig) {
	if obj.Type == "" {
		obj.Type = "dhcp"
	}
}

// ServiceEndpoint is the IP or DNS address and port of a service.
type ServiceEndpoint struct {
	// Addr is the network address at which the service is available. This
	// value may be a FQDN or IP address.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Addr string `json:"addr,omitempty"`

	// Port is the port on which the service is listening.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Port int32 `json:"port,omitempty"`

	// Zone is an IPv6 scoped addressing zone.
	// +optional
	Zone string `json:"zone,omitempty"`
}

// String returns the textual representation of a ServiceEndpoint.
func (a ServiceEndpoint) String() string {
	if a.Addr == "" && a.Port == 0 && a.Zone == "" {
		return ""
	}
	w := &bytes.Buffer{}
	if strings.Contains(a.Addr, ":") {
		fmt.Fprintf(w, "[%s]", a.Addr)
	} else {
		fmt.Fprintf(w, "%s", a.Addr)
	}
	if a.Zone != "" {
		fmt.Fprintf(w, "%%%s", a.Zone)
	}
	if a.Port > 0 {
		fmt.Fprintf(w, ":%d", a.Port)
	}
	return w.String()
}
