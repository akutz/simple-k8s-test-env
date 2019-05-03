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

// Delete the machine. If no error is returned, it is assumed that all
// dependent resources have been cleaned up.
func (a actuator) Delete(
	ctx context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine) error {

	if err := a.delete(ctx, cluster, machine); err != nil {
		msg := err.Error()
		machine.Status.ErrorMessage = &msg
		return err
	}
	return nil
}

func (a actuator) delete(
	parent context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine) error {

	ctx, err := newRequestContext(parent, cluster, machine)
	if err != nil {
		return err
	}
	if err := a.vmDestroy(ctx); err != nil {
		return err
	}
	return nil
}
