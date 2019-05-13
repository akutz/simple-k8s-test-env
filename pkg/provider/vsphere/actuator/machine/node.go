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
	"bytes"
	"encoding/base64"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/ssh"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
)

func (a actuator) nodeEnsure(ctx *reqctx) error {
	if err := a.nodeEnsureJoin(ctx); err != nil {
		return err
	}
	return nil
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
		"sudo sh -c '[ -e /etc/kubernetes/kubelet.conf ]' || "+
			"sudo %s", ctx.csta.KubeJoinCmd)
}

func (a actuator) nodeEnsureCloudProvider(ctx *reqctx) error {
	cloudProvider := ctx.cluster.Labels[config.CloudProviderLabelName]
	switch ccm := ctx.ccfg.CloudProvider.Object.(type) {
	case *vconfig.InternalCloudProviderConfig:
		if cloudProvider == "vsphere" {
			return a.nodeEnsureInternalCloudProvider(ctx, ccm)
		}
	}
	return nil
}

func (a actuator) nodeEnsureInternalCloudProvider(
	ctx *reqctx,
	ccm *vconfig.InternalCloudProviderConfig) error {

	log.WithField("vm", ctx.machine.Name).Info("kube-apply-internal-ccm")

	configDir := ctx.cluster.Labels[config.ConfigDirLabelName]
	if configDir == "" {
		return nil
	}

	sshClient, err := ssh.NewClient(*ctx.msta.SSH, ctx.ccfg.SSH)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	if err := ssh.MkdirAll(
		ctx,
		sshClient,
		"/etc/kubernetes",
		"root", "root", 0755); err != nil {
		return err
	}

	fmap := template.FuncMap{
		"join": strings.Join,
		"base64": func(plain string) string {
			return base64.StdEncoding.EncodeToString([]byte(plain))
		},
	}

	tpl, err := template.New("t").Funcs(fmap).Parse(ccm.Templates.Config)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, ccm); err != nil {
		return err
	}
	if err := ssh.Upload(
		ctx, sshClient,
		buf.Bytes(), ccm.ConfigFilePath,
		"root", "root", 0640); err != nil {
		return err
	}

	return nil
}
