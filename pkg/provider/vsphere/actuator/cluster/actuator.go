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

import (
	"sigs.k8s.io/cluster-api/pkg/controller/cluster"

	"vmware.io/sk8/pkg/provider"
	"vmware.io/sk8/pkg/provider/vsphere/config"
)

func init() {
	provider.RegisterClusterActuator(config.GroupName, actuator{})
}

type actuator struct {
}

// New returns a new vSphere cluster actuator.
func New() cluster.Actuator {
	return actuator{}
}