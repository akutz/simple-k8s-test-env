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
	"net"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/config/encoding"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
)

// NewExpectedClusterProviderConfig returns a new ClusterProviderConfig
// object used for testing.
func NewExpectedClusterProviderConfig() *vconfig.ClusterProviderConfig {

	lvsConfig := &config.LinuxVirtualSwitchConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: config.SchemeGroupVersion.String(),
			Kind:       "LinuxVirtualSwitchConfig",
		},
		PublicIPAddr:  net.ParseIP("1.2.3.4"),
		PrivateIPAddr: net.ParseIP("5.6.7.8"),
		SSH: config.SSHCredentialAndEndpoint{
			SSHEndpoint: config.SSHEndpoint{
				ServiceEndpoint: config.ServiceEndpoint{
					Addr: "lvs.host",
				},
			},
		},
	}
	if d := os.Getenv("SK8_SSH_DIR"); d != "" {
		lvsConfig.SSH.PrivateKeyPath = "id_rsa.sk8"
		lvsConfig.SSH.PublicKeyPath = "id_rsa.sk8.pub"
	}
	encoding.Scheme.Default(lvsConfig)

	obj := vconfig.ClusterProviderConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: vconfig.SchemeGroupVersion.String(),
			Kind:       "ClusterProviderConfig",
		},
		NAT: &runtime.RawExtension{
			Object: lvsConfig,
		},
	}
	if d := os.Getenv("SK8_SSH_DIR"); d != "" {
		obj.SSH.PrivateKeyPath = "id_rsa.sk8"
		obj.SSH.PublicKeyPath = "id_rsa.sk8.pub"
	}
	encoding.Scheme.Default(&obj)
	return &obj
}
