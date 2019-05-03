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

package machine

import (
	"context"

	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// Checks if the machine currently exists.
func (a actuator) Exists(
	ctx context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine) (bool, error) {

	ok, err := a.exists(ctx, cluster, machine)
	if err != nil {
		msg := err.Error()
		machine.Status.ErrorMessage = &msg
		return false, err
	}
	return ok, nil
}

func (a actuator) exists(
	parent context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine) (bool, error) {

	ctx, err := newRequestContext(parent, cluster, machine)
	if err != nil {
		return false, err
	}
	return a.vmExists(ctx)
}
