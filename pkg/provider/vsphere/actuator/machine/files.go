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
	"text/template"

	log "github.com/sirupsen/logrus"

	"vmware.io/sk8/pkg/net/ssh"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
	vkubeadm "vmware.io/sk8/pkg/provider/vsphere/kubeadm"
)

func (a actuator) filesEnsure(ctx *reqctx) error {
	log.WithField("vm", ctx.machine.Name).Debug("files-ensure")
	if err := a.filesEnsureEtcKubernetes(ctx); err != nil {
		return err
	}
	if err := a.filesEnsureKubeadmConf(ctx); err != nil {
		return err
	}
	if err := a.filesEnsureCloudConf(ctx); err != nil {
		return err
	}
	return nil
}

func (a actuator) filesEnsureEtcKubernetes(ctx *reqctx) error {
	log.WithField("vm", ctx.machine.Name).Debug("files-ensure-etc-kubernetes")
	return ssh.MkdirAll(
		ctx,
		ctx.ssh,
		"/etc/kubernetes",
		"root", "root", 0755)
}

const kubeadmConfPath = "/etc/kubernetes/kubeadm.conf"

func (a actuator) filesEnsureKubeadmConf(ctx *reqctx) error {
	log.WithField("vm", ctx.machine.Name).Debug("files-ensure-kubeadm-conf")

	if err := ssh.FileExists(
		ctx, ctx.ssh, kubeadmConfPath); err == nil {
		return nil
	}

	data := vkubeadm.ConfigData{
		ClusterName:       ctx.cluster.Name,
		KubernetesVersion: ctx.machine.Spec.Versions.Kubelet,
		//ControlPlaneEndpoint: net.JoinHostPort(
		//	ctx.cluster.Status.APIEndpoints[0].Host,
		//	strconv.Itoa(ctx.cluster.Status.APIEndpoints[0].Port)),
		APIAdvertiseAddress: ctx.msta.IPAddr,
		APIBindPort:         443,
		APIServerCertSANs:   []string{ctx.cluster.Status.APIEndpoints[0].Host},
		DNSDomain:           ctx.cluster.Spec.ClusterNetwork.ServiceDomain,
	}

	if n := ctx.cluster.Spec.ClusterNetwork; true {
		if v := n.Pods.CIDRBlocks; len(v) > 0 {
			data.PodSubnet = n.Pods.CIDRBlocks[0]
		}
		if v := n.Services.CIDRBlocks; len(v) > 0 {
			data.ServiceSubnet = n.Services.CIDRBlocks[0]
		}
	}

	switch ccm := ctx.ccfg.CloudProvider.Object.(type) {
	case *vconfig.ExternalCloudProviderConfig:
		data.CloudProvider = "external"
	case *vconfig.InternalCloudProviderConfig:
		data.CloudProvider = "vsphere"
		data.CloudConfig = ccm.ConfigFilePath
	}

	buf, err := vkubeadm.Config(data)
	if err != nil {
		return err
	}

	return ssh.Upload(
		ctx, ctx.ssh,
		buf, kubeadmConfPath,
		"root", "root", 0640)
}

func (a actuator) filesEnsureCloudConf(ctx *reqctx) error {
	ccm, ok := ctx.ccfg.CloudProvider.Object.(*vconfig.InternalCloudProviderConfig)
	if !ok {
		return nil
	}

	log.WithField("vm", ctx.machine.Name).Debug("files-ensure-cloud-conf")

	tpl, err := template.New("t").Funcs(tplFuncMap).Parse(ccm.Templates.Config)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, ccm); err != nil {
		return err
	}
	if err := ssh.Upload(
		ctx, ctx.ssh,
		buf.Bytes(), ccm.ConfigFilePath,
		"root", "root", 0640); err != nil {
		return err
	}

	return nil
}
