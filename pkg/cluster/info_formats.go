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

package cluster

// DefaultTemplate is used to emit information about a cluster.
const DefaultTemplate = `Name:          {{.Name}}{{if .Created}}
Created:       {{.Created}}{{end}}{{if .Deleted}}
Deleted:       {{.Deleted}}{{end}}
Kubeconfig:    {{.Kubeconfig}}{{if .CloudProvider}}
CloudProvider: {{.CloudProvider}}{{end}}
Machines:{{range .Machines}}
  Name:        {{.Name}}{{if .Created}}
  Created:     {{.Created}}{{end}}{{if .Deleted}}
  Deleted:     {{.Deleted}}{{end}}
  Roles:       {{.Roles.String}}
  Versions:{{if .IsController}}
    Control:   {{.Versions.ControlPlane}}{{end}}
    Kubelet:   {{.Versions.Kubelet}}{{end}}
`
