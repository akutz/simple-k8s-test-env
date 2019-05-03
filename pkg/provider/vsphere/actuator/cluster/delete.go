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

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/lvs"
	"vmware.io/sk8/pkg/net/ssh"
	vutil "vmware.io/sk8/pkg/provider/vsphere/util"
)

// Delete the cluster.
func (a actuator) Delete(cluster *capi.Cluster) error {
	return a.DeleteWithContext(context.Background(), cluster)
}

// DeleteWithContext the cluster.
func (a actuator) DeleteWithContext(
	ctx context.Context, cluster *capi.Cluster) error {

	// Get the cluster provider config.
	ccfg := vutil.GetClusterProviderConfig(cluster)

	// Check to see if the cluster provider is using a NAT provider.
	switch tnat := ccfg.NAT.Object.(type) {
	case *config.LinuxVirtualSwitchConfig:
		sshClient, err := ssh.NewClient(
			tnat.SSH.SSHEndpoint,
			tnat.SSH.SSHCredential)
		if err != nil {
			return errors.Wrap(err, "error dialing lvs host")
		}
		defer sshClient.Close()
		deleteService := func(sid string) error {
			err := lvs.DeleteTCPService(
				ctx,
				sshClient,
				tnat.PublicNIC,
				sid,
				tnat.PublicIPAddr.String())
			if err != nil {
				return errors.Wrapf(
					err, "error deleting lvs service %q", sid)
			}
			log.WithFields(log.Fields{
				"service": sid,
				"cluster": cluster.Name,
			}).Info("deleted lvs service")
			return nil
		}
		if err := deleteService(cluster.Name + "-api"); err != nil {
			return err
		}
		if err := deleteService(cluster.Name + "-ssh"); err != nil {
			return err
		}
	}

	return nil
}
