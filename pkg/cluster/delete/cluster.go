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

package delete

import (
	"context"
	"fmt"
	//"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"vmware.io/sk8/pkg/cluster"
	"vmware.io/sk8/pkg/provider"
	"vmware.io/sk8/pkg/status"
)

// Cluster turns down an existing Kubernetes cluster.
func Cluster(ctx context.Context, obj *cluster.Cluster) error {
	defer status.End(ctx, false)

	clu := &obj.Cluster
	items := obj.Machines.Items

	fmt.Printf("Deleting cluster %q ...\n", clu.Name)

	status.Start(
		ctx,
		fmt.Sprintf("Deleting %d machines(s) 🖥", len(items)))
	for i := range items {
		machine := &items[i]

		// Get the machine actuator.
		cfg := machine.Spec.ProviderSpec.Value.Object
		gvk := cfg.GetObjectKind().GroupVersionKind()
		act := provider.NewMachineActuator(gvk.Group)

		// Delete the machine.
		log.WithField("name", machine.Name).Debug("deleting machine")
		if err := act.Delete(ctx, clu, machine); err != nil {
			return errors.Wrapf(
				err, "error deleting machine %q", machine.Name)
		}

		machine.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	}

	{
		// Get the cluster actuator.
		cfg := clu.Spec.ProviderSpec.Value.Object
		gvk := cfg.GetObjectKind().GroupVersionKind()
		act := provider.NewClusterActuator(gvk.Group)

		// Delete the cluster.
		log.WithField("name", clu.Name).Debug("deleting cluster")
		status.Start(ctx, "Deleting cluster 🗄")
		var err error
		if act2, ok := act.(cluster.ActuatorWithContext); ok {
			err = act2.DeleteWithContext(ctx, clu)
		} else {
			err = act.Delete(clu)
		}
		if err != nil {
			return errors.Wrapf(
				err, "error deleting cluster %q", clu.Name)
		}
		clu.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	}

	// Delete the cluster's local config directory.
	//if p := cluster.FilePath(clu.Name); p != "" {
	//	log.WithField("path", p).Debug("removing cluster data directory")
	//	os.RemoveAll(p)
	//}
	status.End(ctx, true)

	return nil
}
