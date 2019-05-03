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
	"bytes"
	"context"
	"text/template"

	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"vmware.io/sk8/pkg/util"
)

// GetMetadata returns the cloud-init metadata for a machine.
func GetMetadata(
	ctx context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine,
	networkConfigData []byte) ([]byte, error) {

	encNetCfgData, err := util.Base64GzipBytes(networkConfigData)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	tpl := template.Must(template.New("t").Parse(metadataPatt))
	if err := tpl.Execute(buf, struct {
		NetworkConfig string
		HostFQDN      string
		InstanceID    string
	}{
		NetworkConfig: encNetCfgData,
		HostFQDN:      machine.Name,
		InstanceID:    machine.Name,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const metadataPatt = `{
  "network": "{{.NetworkConfig}}",
  "network.encoding": "gzip+base64",
  "local-hostname": "{{.HostFQDN}}",
  "instance-id": "{{.HostFQDN}}"
}
`
