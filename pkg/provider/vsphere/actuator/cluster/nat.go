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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/lvs"
	"vmware.io/sk8/pkg/net/ssh"
)

func (a actuator) natEnsure(ctx *reqctx) error {
	switch nat := ctx.ccfg.NAT.Object.(type) {
	case *config.LinuxVirtualSwitchConfig:
		return a.natEnsureLVS(ctx, nat)
	}
	return nil
}

func (a actuator) natEnsureLVS(
	ctx *reqctx,
	nat *config.LinuxVirtualSwitchConfig) error {

	sshClient, err := ssh.NewClient(
		nat.SSH.SSHEndpoint,
		nat.SSH.SSHCredential)
	if err != nil {
		return errors.Wrap(err, "error dialing lvs host")
	}
	defer sshClient.Close()
	createService := func(sid string) (int, error) {
		port, err := lvs.CreateRoundRobinTCPService(
			ctx,
			sshClient,
			nat.PublicNIC,
			sid,
			nat.PublicIPAddr.String())
		if err != nil {
			return 0, errors.Wrapf(
				err, "error creating lvs service %q", sid)
		}
		log.WithFields(log.Fields{
			"service": sid,
			"port":    port,
			"cluster": ctx.cluster.Name,
		}).Info("created lvs service")
		return port, nil
	}
	if len(ctx.cluster.Status.APIEndpoints) == 0 {
		port, err := createService(ctx.cluster.Name + "-api")
		if err != nil {
			return err
		}
		ctx.cluster.Status.APIEndpoints = append(
			ctx.cluster.Status.APIEndpoints,
			capi.APIEndpoint{
				Host: nat.SSH.Addr,
				Port: port,
			})
	}
	if ctx.csta.SSH == nil {
		port, err := createService(ctx.cluster.Name + "-ssh")
		if err != nil {
			return err
		}
		ctx.csta.SSH = &config.SSHEndpoint{
			ServiceEndpoint: config.ServiceEndpoint{
				Addr: nat.SSH.Addr,
				Port: int32(port),
			},
		}
	}

	return nil
}
