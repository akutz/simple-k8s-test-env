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
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/lvs"
	"vmware.io/sk8/pkg/net/ssh"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
	"vmware.io/sk8/pkg/status"
)

func (a actuator) apiEnsure(ctx *reqctx) error {
	if ctx.role.Has(config.MachineRoleControlPlane) {
		if err := a.apiEnsureAccess(ctx); err != nil {
			return err
		}
		// If the machine is a member of the control plane then either bring
		// the control plane online or wait for the control plane to come
		// online.
		//
		// Machines that are not members of the control plane can only wait
		// for the control plane to come online.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ctx.csta.ControlPlaneCanOwn:
			defer status.End(ctx, false)
			status.Start(ctx, "Configuring control plane 👀")
			if err := a.apiEnsureInit(ctx); err != nil {
				return err
			}
			if err := a.apiEnsureKubeConf(ctx); err != nil {
				return err
			}
			if err := a.apiEnsureNetwork(ctx); err != nil {
				return err
			}
			if err := a.apiEnsureCloudProvider(ctx); err != nil {
				return err
			}
			status.End(ctx, true)
		case <-ctx.csta.ControlPlaneOnline:
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ctx.csta.ControlPlaneOnline:
	}

	log.WithField("vm", ctx.machine.Name).Debug("control plane online")
	return nil
}

func (a actuator) apiEnsureAccess(ctx *reqctx) error {
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
			Addr: ctx.msta.IPAddr,
			Port: 443,
		}
		if lvsTgt.Addr == "" {
			return errors.Errorf("no LVS target %q", ctx.machine.Name)
		}

		// Either set or get the target of the LVS SSH service.
		if err := lvs.AddTargetToRoundRobinTCPService(
			ctx,
			sshClient,
			ctx.cluster.Name+"-api",
			nat.PublicIPAddr.String(),
			lvsTgt); err != nil {
			return err
		}
	}

	return nil
}

func (a actuator) apiEnsureInit(ctx *reqctx) error {
	log.WithField("vm", ctx.machine.Name).Debug("kubeadm-init")

	if ssh.FileExists(ctx, ctx.ssh, "/etc/kubernetes/admin.conf") == nil {
		log.WithField("vm", ctx.machine.Name).Info("kubeadm-init-idempotent")
		return nil
	}

	ctx.csta.ControlPlaneEndpoint = net.JoinHostPort(ctx.msta.IPAddr, "443")
	log.WithField(
		"ipAddr",
		ctx.csta.ControlPlaneEndpoint).Info("control-plane-endpoint")

	log.WithField("vm", ctx.machine.Name).Info("kubeadm-init-new")
	cmd := "sudo kubeadm init --config " + kubeadmConfPath

	stdout := &bytes.Buffer{}
	fmt.Fprintln(stdout, cmd)
	if err := ssh.Run(ctx, ctx.ssh, nil, stdout, nil, cmd); err != nil {
		return err
	}

	kubeInitLog := path.Join(ctx.dir, "kubeadm-init.log")
	ioutil.WriteFile(kubeInitLog, stdout.Bytes(), 0640)

	re := regexp.MustCompile(`^\s*(kubeadm join.*)$`)
	joinBuf := &bytes.Buffer{}
	scn := bufio.NewScanner(stdout)
	for scn.Scan() {
		line := scn.Text()
		if m := re.FindStringSubmatch(line); len(m) > 0 {
			line = m[1]
			multiLine := strings.HasSuffix(line, `\`)
			joinBuf.WriteString(strings.TrimSuffix(line, `\`))
			for multiLine && scn.Scan() {
				line = scn.Text()
				multiLine = strings.HasSuffix(line, `\`)
				line = strings.TrimSpace(line)
				joinBuf.WriteString(strings.TrimSuffix(line, `\`))
			}
			break
		}
	}

	ctx.csta.KubeJoinCmd = joinBuf.String() + " --config " + kubeadmConfPath
	close(ctx.csta.ControlPlaneOnline)

	return nil
}

func (a actuator) apiEnsureNetwork(ctx *reqctx) error {
	if len(ctx.cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) > 0 {
		return nil
	}
	log.WithField("vm", ctx.machine.Name).Info("kube-apply-networking")

	return ssh.Run(
		ctx, ctx.ssh,
		strings.NewReader(ctx.ccfg.Net),
		nil, nil,
		sudoKubectlApplyStdin)
}

func (a actuator) apiEnsureKubeConf(ctx *reqctx) error {
	log.WithField("vm", ctx.machine.Name).Info("kube-config-get")
	stdout := &bytes.Buffer{}
	if err := ssh.Run(ctx, ctx.ssh, nil, stdout, nil,
		"sudo cat /etc/kubernetes/admin.conf"); err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(stdout)
	if err != nil {
		return errors.Wrap(err, "error reading remote kubeConfig")
	}

	// Replace the server URL with the external IP.
	re := regexp.MustCompile(`server:\shttps://[^\s]+`)
	newAPIServer := []byte(fmt.Sprintf("server: https://%s:%d",
		ctx.cluster.Status.APIEndpoints[0].Host,
		ctx.cluster.Status.APIEndpoints[0].Port))
	kubeConfig := re.ReplaceAllLiteral(buf, newAPIServer)
	kubeConfigPath := path.Join(ctx.dir, "kube.conf")
	if err := ioutil.WriteFile(kubeConfigPath, kubeConfig, 0640); err != nil {
		return errors.Wrapf(err, "error writing kubeConfig %q", kubeConfigPath)
	}

	return nil
}

func (a actuator) apiEnsureCloudProvider(ctx *reqctx) error {
	switch ccm := ctx.ccfg.CloudProvider.Object.(type) {
	case *vconfig.ExternalCloudProviderConfig:
		return a.apiEnsureExternalCloudProvider(ctx, ccm)
	case *vconfig.InternalCloudProviderConfig:
		return a.apiEnsureInternalCloudProvider(ctx, ccm)
	}
	return nil
}

func (a actuator) apiEnsureExternalCloudProvider(
	ctx *reqctx,
	ccm *vconfig.ExternalCloudProviderConfig) error {

	log.WithField("vm", ctx.machine.Name).Info("kube-apply-external-ccm")
	kubeApply := func(format, objectName string) error {
		tpl, err := template.New("t").Funcs(tplFuncMap).Parse(format)
		if err != nil {
			return err
		}
		buf := &bytes.Buffer{}
		if err := tpl.Execute(buf, ccm); err != nil {
			return err
		}

		yamlFile := path.Join(
			ctx.dir,
			fmt.Sprintf("ext-ccm-%s.yaml", objectName))
		if err := ioutil.WriteFile(yamlFile, buf.Bytes(), 0640); err != nil {
			return err
		}

		if err := ssh.Run(
			ctx,
			ctx.ssh,
			buf,
			nil,
			nil,
			sudoKubectlApplyStdin); err != nil {
			return errors.Wrapf(err, "error creating %s", objectName)
		}

		return nil
	}

	// 1. Create the ConfigMap
	if err := kubeApply(ccm.Templates.Config, "ConfigMap"); err != nil {
		return err
	}
	// 2. Create the ServiceAccount
	if err := kubeApply(ccm.Templates.ServiceAccount, "ServiceAccount"); err != nil {
		return err
	}
	// 3. Create the Roles
	if err := kubeApply(ccm.Templates.Roles, "Roles"); err != nil {
		return err
	}
	// 4. Create the RoleBindings
	if err := kubeApply(ccm.Templates.RoleBindings, "RoleBindings"); err != nil {
		return err
	}
	// 5. Create the Secrets
	if err := kubeApply(ccm.Templates.Secrets, "Secrets"); err != nil {
		return err
	}
	// 6. Create the Service
	if err := kubeApply(ccm.Templates.Service, "Service"); err != nil {
		return err
	}
	// 7. Create the Deployment
	if err := kubeApply(ccm.Templates.Deployment, "Deployment"); err != nil {
		return err
	}

	return nil
}

func (a actuator) apiEnsureInternalCloudProvider(
	ctx *reqctx,
	ccm *vconfig.InternalCloudProviderConfig) error {

	log.WithField("vm", ctx.machine.Name).Info("kube-apply-internal-ccm")
	kubeApply := func(format, objectName string) error {
		tpl, err := template.New("t").Funcs(tplFuncMap).Parse(format)
		if err != nil {
			return err
		}
		buf := &bytes.Buffer{}
		if err := tpl.Execute(buf, ccm); err != nil {
			return err
		}

		yamlFile := path.Join(
			ctx.dir,
			fmt.Sprintf("int-ccm-%s.yaml", objectName))
		if err := ioutil.WriteFile(yamlFile, buf.Bytes(), 0640); err != nil {
			return err
		}

		if err := ssh.Run(
			ctx,
			ctx.ssh,
			buf,
			nil,
			nil,
			sudoKubectlApplyStdin); err != nil {
			return errors.Wrapf(err, "error creating %s", objectName)
		}

		return nil
	}

	// 1. Create the Secrets
	if err := kubeApply(ccm.Templates.Secrets, "Secrets"); err != nil {
		return err
	}

	return nil
}
