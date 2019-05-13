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

package config_test

import (
	"testing"

	"sigs.k8s.io/yaml"

	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
)

func TestInternalCloudProviderConfig(t *testing.T) {
	cfg := &vconfig.InternalCloudProviderConfig{}
	vconfig.SetObjectDefaults_InternalCloudProviderConfig(cfg)
	buf, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(buf))
}
