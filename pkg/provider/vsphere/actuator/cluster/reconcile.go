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
	"context"

	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// Reconcile creates or applies updates to the cluster.
func (a actuator) Reconcile(cluster *capi.Cluster) error {
	return a.ReconcileWithContext(context.Background(), cluster)
}

// Reconcile creates or applies updates to the cluster.
func (a actuator) ReconcileWithContext(
	ctx context.Context, cluster *capi.Cluster) error {

	return a.reconcileWithContext(ctx, cluster)
}

func (a actuator) reconcileWithContext(
	parent context.Context,
	cluster *capi.Cluster) error {

	ctx, err := newRequestContext(parent, cluster)
	if err != nil {
		return err
	}
	if err := a.ccmEnsure(ctx); err != nil {
		return err
	}
	if err := a.ovaEnsure(ctx); err != nil {
		return err
	}
	if err := a.natEnsure(ctx); err != nil {
		return err
	}
	if err := a.sshEnsure(ctx); err != nil {
		return err
	}
	return nil
}
