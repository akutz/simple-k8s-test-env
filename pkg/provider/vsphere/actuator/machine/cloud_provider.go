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
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/ssh"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
)

func (a actuator) ccmEnsure(ctx *reqctx) error {
	cloudProvider := ctx.cluster.Labels[config.CloudProviderLabelName]
	switch ccm := ctx.ccfg.CloudProvider.Object.(type) {
	case *vconfig.ExternalCloudProviderConfig:
		if cloudProvider == "external" {
			return a.ccmEnsureExt(ctx, ccm)
		}
	}
	return nil
}

func (a actuator) ccmEnsureExt(
	ctx *reqctx,
	ccm *vconfig.ExternalCloudProviderConfig) error {

	log.WithField("vm", ctx.machine.Name).Info("kube-apply-external-ccm")

	configDir := ctx.cluster.Labels[config.ConfigDirLabelName]
	if configDir == "" {
		return nil
	}

	sshClient, err := ssh.NewClient(*ctx.msta.SSH, ctx.ccfg.SSH)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	fmap := template.FuncMap{
		"join": strings.Join,
		"base64": func(plain string) string {
			return base64.StdEncoding.EncodeToString([]byte(plain))
		},
	}

	kubeApply := func(format, objectName string) error {
		tpl, err := template.New("t").Funcs(fmap).Parse(format)
		if err != nil {
			return err
		}
		buf := &bytes.Buffer{}
		if err := tpl.Execute(buf, ccm); err != nil {
			return err
		}

		yamlFile := path.Join(
			configDir,
			fmt.Sprintf("ext-ccm-%s.yaml", objectName))
		if err := ioutil.WriteFile(yamlFile, buf.Bytes(), 0640); err != nil {
			return err
		}

		if err := ssh.Run(
			ctx,
			sshClient,
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
