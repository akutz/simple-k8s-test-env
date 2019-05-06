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

	log "github.com/sirupsen/logrus"
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// Delete the cluster.
func (a actuator) Delete(cluster *capi.Cluster) error {
	return a.DeleteWithContext(context.Background(), cluster)
}

// DeleteWithContext the cluster.
func (a actuator) DeleteWithContext(
	ctx context.Context, cluster *capi.Cluster) error {

	return a.deleteWithContext(ctx, cluster)
}

func (a actuator) deleteWithContext(
	parent context.Context, cluster *capi.Cluster) error {

	ctx, err := newRequestContext(parent, cluster)
	if err != nil {
		return err
	}
	if err := a.natDelete(ctx); err != nil {
		log.WithError(err).Debug("error delting nat resources")
	}
	return nil
}
