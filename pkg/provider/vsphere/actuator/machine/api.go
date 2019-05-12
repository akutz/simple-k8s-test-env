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
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/lvs"
	"vmware.io/sk8/pkg/net/ssh"
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
			if err := a.ccmEnsure(ctx); err != nil {
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
			Port: 443,
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
	log.WithField("vm", ctx.machine.Name).Info("kubeadm-init")
	sshClient, err := ssh.NewClient(*ctx.msta.SSH, ctx.ccfg.SSH)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	cmd := &bytes.Buffer{}
	fmt.Fprintf(
		cmd,
		"sudo kubeadm init --apiserver-bind-port=443"+
			" --kubernetes-version=%q",
		ctx.machine.Spec.Versions.ControlPlane)

	for _, e := range ctx.cluster.Status.APIEndpoints {
		fmt.Fprintf(cmd, " --apiserver-cert-extra-sans=%q", e.Host)
	}
	if n := ctx.cluster.Spec.ClusterNetwork; true {
		if v := n.Pods.CIDRBlocks; len(v) > 0 {
			fmt.Fprintf(cmd, " --pod-network-cidr=%q", v[0])
		}
		if v := n.Services.CIDRBlocks; len(v) > 0 {
			fmt.Fprintf(cmd, " --service-cidr=%q", v[0])
		}
		if v := n.ServiceDomain; v != "" {
			fmt.Fprintf(cmd, " --service-dns-domain=%q", v)
		}
	}

	stdout := &bytes.Buffer{}
	fmt.Fprintln(stdout, cmd.String())
	if err := ssh.Run(
		ctx, sshClient, nil, stdout, nil, cmd.String()); err != nil {
		return err
	}

	configDir := ctx.cluster.Labels[config.ConfigDirLabelName]
	if configDir == "" {
		return nil
	}

	kubeInitLog := path.Join(configDir, "kubeadm-init.log")
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

	ctx.csta.KubeJoinCmd = joinBuf.String()
	close(ctx.csta.ControlPlaneOnline)

	return nil
}

func (a actuator) apiEnsureNetwork(ctx *reqctx) error {
	if len(ctx.cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) > 0 {
		return nil
	}
	log.WithField("vm", ctx.machine.Name).Info("kube-apply-networking")

	sshClient, err := ssh.NewClient(*ctx.msta.SSH, ctx.ccfg.SSH)
	if err != nil {
		return err
	}
	defer sshClient.Close()
	return ssh.Run(
		ctx, sshClient,
		strings.NewReader(ctx.ccfg.Net),
		nil, nil,
		sudoKubectlApplyStdin)
}

func (a actuator) apiEnsureKubeConf(ctx *reqctx) error {
	configDir := ctx.cluster.Labels[config.ConfigDirLabelName]
	if configDir == "" {
		return nil
	}

	log.WithField("vm", ctx.machine.Name).Info("kube-config-get")
	sshClient, err := ssh.NewClient(*ctx.msta.SSH, ctx.ccfg.SSH)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	stdout := &bytes.Buffer{}
	if err := ssh.Run(ctx, sshClient, nil, stdout, nil,
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
	kubeConfigPath := path.Join(configDir, "kube.conf")
	if err := ioutil.WriteFile(kubeConfigPath, kubeConfig, 0640); err != nil {
		return errors.Wrapf(err, "error writing kubeConfig %q", kubeConfigPath)
	}

	return nil
}
