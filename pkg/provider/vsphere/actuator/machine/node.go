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
	log "github.com/sirupsen/logrus"

	"vmware.io/sk8/pkg/net/ssh"
)

func (a actuator) nodeEnsure(ctx *reqctx) error {
	return a.nodeEnsureJoin(ctx)
}

func (a actuator) nodeEnsureJoin(ctx *reqctx) error {
	log.WithField("vm", ctx.machine.Name).Info("kubeadm-join")
	sshClient, err := ssh.NewClient(*ctx.msta.SSH, ctx.ccfg.SSH)
	if err != nil {
		return err
	}
	defer sshClient.Close()
	return ssh.Run(
		ctx,
		sshClient,
		nil, nil, nil,
		"sh -c '[ -d /etc/kubernetes ] || sudo %s'", ctx.csta.KubeJoinCmd)
}
