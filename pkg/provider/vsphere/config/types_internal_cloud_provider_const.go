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

const intCCMConfigFormat = `[Global]
datacenters        = "{{.Datacenter}}"
insecure-flag      = "{{.Insecure}}"
port               = "{{.ServerPort}}"
secret-name        = "{{.SecretName}}"
secret-namespace   = "{{.Namespace}}"

[VirtualCenter "{{.ServerAddr}}"]

[Workspace]
server             = "{{.ServerAddr}}"
datacenter         = "{{.Datacenter}}"
folder             = "{{.Folder}}"
default-datastore  = "{{.Datastore}}"
resourcepool-path  = "{{.ResourcePool}}"

[Disk]
scsicontrollertype = {{.SCSIControllerType}}

[Network]
public-network     = "{{.Network}}"{{if .Region}}{{if .Zone}}

[Labels]
region = "{{.Region}}"
zone =   "{{.Zone}}"{{end}}{{end}}
`

const intCCMSecretsFormat = `apiVersion: v1
kind: Secret
metadata:
  name: {{.SecretName}}
  namespace: {{.Namespace}}
data:
  {{.ServerAddr}}.username: "{{base64 .Username}}"
  {{.ServerAddr}}.password: "{{base64 .Password}}"
`
