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

package test_test

import (
	"fmt"
	"net"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"

	"vmware.io/sk8/pkg/cluster"
	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/config/encoding"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
	vtest "vmware.io/sk8/pkg/provider/vsphere/test"
	"vmware.io/sk8/pkg/test"
)

func TestMain(m *testing.M) {
	test.UpdateLogLevel()
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Setenv("SK8_SSH_DIR", path.Join(wd, "data"))
	os.Setenv("VSPHERE_SERVER", "env:VSPHERE_SERVER")
	os.Setenv("VSPHERE_USERNAME", "env:VSPHERE_USERNAME")
	os.Setenv("VSPHERE_PASSWORD", "env:VSPHERE_PASSWORD")
	os.Exit(m.Run())
}

func TestReadFromFile(t *testing.T) {
	cmpEqEmpty := cmpopts.EquateEmpty()
	cmpIgnoreSSHPrvKeyPath := cmpopts.IgnoreFields(
		config.SSHCredential{},
		"PrivateKeyPath")
	cmpIgnoreSSHPubKeyPath := cmpopts.IgnoreFields(
		config.SSHCredential{},
		"PublicKeyPath")

	expObj, err := vtest.NewExpectedCluster()
	if err != nil {
		t.Fatalf("failed to get exp cluster: %v", err)
	}

	cases := []struct {
		TestName    string
		Path        string
		ExpectError bool
	}{
		{
			TestName:    "No config",
			Path:        "",
			ExpectError: true,
		},
		{
			TestName:    "Invalid path",
			Path:        "./data/not-a-file.bogus",
			ExpectError: true,
		},
		{
			TestName:    "Invalid API version",
			Path:        "./data/invalid-apiversion.yaml",
			ExpectError: true,
		},
		{
			TestName:    "Invalid cluster",
			Path:        "./data/invalid-cluster.yaml",
			ExpectError: true,
		},
		{
			TestName:    "Invalid YAML",
			Path:        "./data/invalid-yaml.yaml",
			ExpectError: true,
		},
		{
			TestName:    "Valid cluster",
			Path:        "./data/valid-cluster.yaml",
			ExpectError: false,
		},
	}
	for _, c := range cases {
		t.Run(c.TestName, func(t *testing.T) {
			handleTestCase := func() error {
				{
					actObj, err := cluster.ReadFromFile(c.Path)
					if err != nil {
						return err
					}
					if _, err := actObj.WithDefaults(
						"./data/valid-defaults.yaml"); err != nil {
						return err
					}

					// Load the actObj's NAT config.
					clusterProviderConfig := actObj.Cluster.Spec.ProviderSpec.Value.Object.(*vconfig.ClusterProviderConfig)
					if _, err := encoding.FromRaw(clusterProviderConfig.NAT); err != nil {
						return err
					}

					if diff := cmp.Diff(
						expObj, actObj,
						cmpEqEmpty,
						cmpIgnoreSSHPrvKeyPath,
						cmpIgnoreSSHPubKeyPath); diff != "" {
						return errors.Errorf("1st diff\n\n%s\n", diff)
					}
				}

				// Create a copy of the expected object in order to update
				// some of its settings.
				expObj2 := expObj.DeepCopy()

				// Update the expObj2's ClusterProviderConfig.
				clusterProviderConfig := expObj2.Cluster.Spec.ProviderSpec.Value.Object.(*vconfig.ClusterProviderConfig)
				lvsConfig := clusterProviderConfig.NAT.Object.(*config.LinuxVirtualSwitchConfig)
				lvsConfig.PublicIPAddr = net.ParseIP("0.0.0.0")

				// Update the expObj2's MachineProviderConfigs.
				for i := range expObj2.Machines.Items {
					m := &expObj2.Machines.Items[i]
					machineProviderConfig := m.Spec.ProviderSpec.Value.Object.(*vconfig.MachineProviderConfig)
					machineProviderConfig.Datacenter = "/SDDC-Datacenter2"
					machineProviderConfig.Network.Interfaces = append(
						machineProviderConfig.Network.Interfaces,
						config.NetworkInterfaceConfig{
							Name: "eth2",
						})
				}

				{
					expObj := expObj2

					actObj, err := cluster.ReadFromFile(c.Path)
					if err != nil {
						return err
					}
					if _, err := actObj.WithDefaults(
						"./data/valid-defaults.yaml",
						"./data/valid-defaults2.yaml"); err != nil {
						return err
					}

					// Load the actObj's NAT config.
					clusterProviderConfig := actObj.Cluster.Spec.ProviderSpec.Value.Object.(*vconfig.ClusterProviderConfig)
					if _, err := encoding.FromRaw(clusterProviderConfig.NAT); err != nil {
						return err
					}

					if diff := cmp.Diff(
						expObj, actObj,
						cmpEqEmpty,
						cmpIgnoreSSHPrvKeyPath,
						cmpIgnoreSSHPubKeyPath); diff != "" {
						return errors.Errorf("2nd diff\n\n%s\n", diff)
					}
				}

				return nil
			}
			if err := handleTestCase(); err != nil && !c.ExpectError {
				t.Fatalf("unexpected error while loading obj: %v", err)
			} else if err == nil && c.ExpectError {
				t.Fatalf("unexpected lack or error while loading obj")
			}
		})
	}
}
