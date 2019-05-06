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

package create

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"vmware.io/sk8/pkg/cluster"
	"vmware.io/sk8/pkg/provider"
	"vmware.io/sk8/pkg/status"
)

// Cluster turns down an existing Kubernetes cluster.
func Cluster(ctx context.Context, obj *cluster.Cluster) error {
	defer status.End(ctx, false)

	clu := &obj.Cluster
	fmt.Printf("Creating cluster %q ...\n", clu.Name)
	{
		// Get the cluster actuator.
		cfg := clu.Spec.ProviderSpec.Value.Object
		gvk := cfg.GetObjectKind().GroupVersionKind()
		act := provider.NewClusterActuator(gvk.Group)

		status.Start(ctx, "Verifying prerequisites 🎈")
		var err error
		if act2, ok := act.(cluster.ActuatorWithContext); ok {
			err = act2.ReconcileWithContext(ctx, clu)
		} else {
			err = act.Reconcile(clu)
		}
		if err != nil {
			return errors.Wrapf(
				err, "error creating cluster %q", clu.Name)
		}
		clu.CreationTimestamp.Time = time.Now()
		status.End(ctx, true)
	}

	status.Start(
		ctx,
		fmt.Sprintf(
			"Creating %d machines(s) 🖥",
			len(obj.Machines.Items)))

	var (
		errs = make(chan error)
		done = make(chan struct{})
		wait sync.WaitGroup
	)

	wait.Add(len(obj.Machines.Items))
	go func() {
		wait.Wait()
		close(done)
	}()

	for i := range obj.Machines.Items {
		machine := &obj.Machines.Items[i]

		// Get the machine actuator.
		cfg := machine.Spec.ProviderSpec.Value.Object
		gvk := cfg.GetObjectKind().GroupVersionKind()
		act := provider.NewMachineActuator(gvk.Group)

		log.WithFields(log.Fields{
			"cluster": clu.Name,
			"machine": machine.Name,
		}).Info("creating machine")

		go func() {
			defer wait.Done()
			err := act.Create(ctx, clu, machine)
			if err != nil {
				errs <- errors.Wrapf(
					err,
					"error creating machine %q for cluster %q",
					machine.Name, clu.Name)
			}
			machine.CreationTimestamp.Time = time.Now()
		}()
	}

	select {
	case <-done:
	case err := <-errs:
		if err != nil {
			return err
		}
	}

	status.End(ctx, true)
	return nil
}
