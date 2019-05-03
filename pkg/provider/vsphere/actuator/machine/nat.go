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
	"github.com/pkg/errors"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/lvs"
	"vmware.io/sk8/pkg/net/ssh"
)

func (a actuator) natEnsure(ctx *reqctx) error {
	// Configure the NAT provider.
	switch nat := ctx.ccfg.NAT.Object.(type) {
	case *config.LinuxVirtualSwitchConfig:
		sshClient, err := ssh.NewClient(
			nat.SSH.SSHEndpoint,
			nat.SSH.SSHCredential)
		if err != nil {
			return err
		}
		defer sshClient.Close()

		// Get the LVS target interface for this machine.
		lvsTgt := config.ServiceEndpoint{
			Port: 22,
		}
		for _, addr := range ctx.machine.Status.Addresses {
			if addr.Type == lvs.NodeIP {
				lvsTgt.Addr = addr.Address
			}
		}
		if lvsTgt.Addr == "" {
			return errors.Errorf("no LVS target %q", ctx.machine.Name)
		}

		// Either set or get the target of the LVS SSH service.
		sshTgt, err := lvs.SetOrGetTargetTCPService(
			ctx,
			sshClient,
			ctx.cluster.Name+"-ssh",
			nat.PublicIPAddr.String(),
			lvsTgt)
		if err != nil {
			return err
		}

		// If the returned SSH target is *this* machine then the
		// machine's SSH endpoint config does not require a bastion
		// proxy. Otherwise it does.
		if *sshTgt == lvsTgt {
			ctx.msta.SSH = ctx.csta.SSH
		} else {
			ctx.msta.SSH = &config.SSHEndpoint{
				ServiceEndpoint: lvsTgt,
				ProxyHost:       ctx.csta.SSH,
			}
		}
	default:
		ctx.msta.SSH = &config.SSHEndpoint{
			ServiceEndpoint: config.ServiceEndpoint{
				Addr: ctx.machine.Status.Addresses[0].Address,
				Port: 22,
			},
		}
	}

	return nil
}
