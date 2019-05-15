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
	log.WithField("vm", ctx.machine.Name).Debug("node-ensure")
	if err := a.nodeEnsureJoin(ctx); err != nil {
		return err
	}
	return nil
}

func (a actuator) nodeEnsureJoin(ctx *reqctx) error {
	log.WithField("vm", ctx.machine.Name).Debug("kubeadm-join")

	if ssh.FileExists(ctx, ctx.ssh, "/etc/kubernetes/kubelet.conf") == nil {
		log.WithField("vm", ctx.machine.Name).Info("kubeadm-join-idempotent")
		return nil
	}

	log.WithField("vm", ctx.machine.Name).Info("kubeadm-join-new")
	return ssh.Run(
		ctx,
		ctx.ssh,
		nil, nil, nil,
		"sudo %s", ctx.csta.KubeJoinCmd)
}
