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
	"strings"
	"text/template"

	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"vmware.io/sk8/pkg/config"
)

// GetNetworkConfig returns the cloud-init network configuration for a machine.
func GetNetworkConfig(
	ctx context.Context,
	machine *capi.Machine,
	machineNetCfg config.MachineNetworkConfig) ([]byte, error) {

	var domain string
	domainParts := strings.SplitN(machine.Name, ".", 2)
	if len(domainParts) > 1 {
		domain = domainParts[1]
	}

	buf := &bytes.Buffer{}
	tpl := template.Must(template.New("t").Parse(networkConfigPatt))
	if err := tpl.Execute(buf, struct {
		config.MachineNetworkConfig
		DomainFQDN string
	}{
		MachineNetworkConfig: machineNetCfg,
		DomainFQDN:           domain,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const networkConfigPatt = `version: 1
config:{{range .Interfaces}}
- type: physical
  name: {{.Name}}
  subnets:
  - type:    {{.Type}}{{if .IPAddr}}
    address: {{.IPAddr}}{{end}}{{if .GatewayIPAddr}}
    gateway: {{.GatewayIPAddr}}{{end}}{{end}}
- type: nameserver
  address:{{range .Nameservers}}
  - {{.}}{{end}}
  search:
  - {{.DomainFQDN}}
`
