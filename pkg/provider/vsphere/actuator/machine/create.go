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

// Create the machine.
func (a actuator) Create(
	ctx context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine) error {

	if err := a.create(ctx, cluster, machine); err != nil {
		msg := err.Error()
		machine.Status.ErrorMessage = &msg
		return err
	}
	return nil
}

func (a actuator) create(
	parent context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine) error {

	ctx, err := newRequestContext(parent, cluster, machine)
	if err != nil {
		return err
	}
	if err := a.vmEnsure(ctx); err != nil {
		return err
	}
	if err := a.natEnsure(ctx); err != nil {
		return err
	}
	if err := a.sshEnsure(ctx); err != nil {
		return err
	}
	defer ctx.ssh.Close()
	if err := a.filesEnsure(ctx); err != nil {
		return err
	}
	if err := a.apiEnsure(ctx); err != nil {
		return err
	}
	if err := a.nodeEnsure(ctx); err != nil {
		return err
	}
	return nil
}
