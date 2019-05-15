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

package kubeadm_test

import (
	"testing"

	"vmware.io/sk8/pkg/provider/vsphere/kubeadm"
)

func TestConfig(t *testing.T) {
	cases := []struct {
		Data        kubeadm.ConfigData
		ExpectError bool
	}{
		{
			Data: kubeadm.ConfigData{
				ClusterName:          "1234567.sk8",
				KubernetesVersion:    "v1.11.0",
				ControlPlaneEndpoint: "192.168.20.54:839",
				APIAdvertiseAddress:  "192.168.20.33",
				APIBindPort:          443,
				APIServerCertSANs:    []string{"192.168.20.54"},
				DNSDomain:            "cluster.local",
				CloudProvider:        "vsphere",
				CloudConfig:          "/etc/kubernetes/cloud.conf",
			},
		},
		{
			Data: kubeadm.ConfigData{
				ClusterName:          "1234567.sk8",
				KubernetesVersion:    "v1.12.0",
				ControlPlaneEndpoint: "192.168.20.54:839",
				APIAdvertiseAddress:  "192.168.20.33",
				APIBindPort:          443,
				APIServerCertSANs:    []string{"192.168.20.54"},
				DNSDomain:            "cluster.local",
				CloudProvider:        "vsphere",
				CloudConfig:          "/etc/kubernetes/cloud.conf",
			},
		},
		{
			Data: kubeadm.ConfigData{
				ClusterName:          "1234567.sk8",
				KubernetesVersion:    "v1.13.2",
				ControlPlaneEndpoint: "192.168.20.54:839",
				APIAdvertiseAddress:  "192.168.20.33",
				APIBindPort:          443,
				APIServerCertSANs:    []string{"192.168.20.54"},
				DNSDomain:            "cluster.local",
				CloudProvider:        "external",
				//CloudConfig:   "/etc/kubernetes/cloud.conf",
			},
		},
		{
			Data: kubeadm.ConfigData{
				ClusterName:          "1234567.sk8",
				KubernetesVersion:    "v1.14.1",
				ControlPlaneEndpoint: "192.168.20.54:839",
				APIAdvertiseAddress:  "192.168.20.33",
				APIBindPort:          443,
				//APIServerCertSANs:    []string{"192.168.20.54"},
				DNSDomain:     "cluster.local",
				CloudProvider: "vsphere",
				CloudConfig:   "/etc/kubernetes/cloud.conf",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Data.KubernetesVersion, func(t *testing.T) {
			config, err := kubeadm.Config(c.Data)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%s", config)
		})
	}

}
