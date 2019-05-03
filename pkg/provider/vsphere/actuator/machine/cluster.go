// +build none

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
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/pkg/errors"

	"vmware.io/sk8/pkg/cluster/access"
	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/config/encoding"
	"vmware.io/sk8/pkg/status"
	"vmware.io/sk8/pkg/util"
	"vmware.io/sk8/pkg/vsphere"
)

// Cluster turns up a new Kubernetes cluster. Information about the cluster
// is added to the provided cluster configuration object.
func Cluster(ctx context.Context, cfg *config.Cluster) error {
	defer status.End(ctx, false)

	// Ensure defaults are set up correctly.
	encoding.Scheme.Default(cfg)

	fmt.Printf("Creating cluster %q ...\n", cfg.ID)

	// Verify that a vSphere client can be created.
	status.Start(ctx, "Verifying vSphere credentials 🔐")
	client, err := vsphere.NewClient(ctx, cfg.VSphere)
	if err != nil {
		return errors.Wrap(err, "error creating vSphere client")
	}
	status.End(ctx, true)

	status.Start(ctx, "Verifying prerequisites 🎈")
	if err := resolveBuildID(ctx, cfg); err != nil {
		return err
	}
	ovaLibItemID, err := client.EnsureLibraryOVA(ctx, cfg.VSphere)
	if err != nil {
		return errors.Wrap(err, "error ensuring library OVA")
	}
	cfg.VSphere.ImportSourceID = ovaLibItemID

	accessProvider, err := access.New(ctx, cfg, true)
	if err != nil {
		return errors.Wrap(err, "error creating external access provider")
	}
	if err := ensureSSHKeyPair(cfg); err != nil {
		return errors.Wrap(err, "error generating ssh keys")
	}
	status.End(ctx, true)

	// Deploy the node VMs in parallel and wait.
	status.Start(ctx,
		fmt.Sprintf("Deploying %d node(s) (1-2 minutes) 🖥", len(cfg.Nodes)))
	if err := vmDeployAll(ctx, cfg, client); err != nil {
		return errors.Wrap(err, "error deploying node VMs")
	}
	// Enable external access.
	if err := accessProvider.Up(ctx); err != nil {
		return errors.Wrap(err, "error configuring external access")
	}
	if err := ensureSSHConfig(cfg); err != nil {
		return errors.Wrap(err, "error creating local ssh config")
	}
	status.End(ctx, true)

	// Wait until the cluster is accessible via SSH.
	status.Start(ctx, "Verifying external access (2-3 minutes) ️️📡️")
	if err := accessProvider.Wait(ctx); err != nil {
		return errors.Wrap(err, "error waiting for external access")
	}
	if err := sshWaitUntilOnline(ctx, *cfg); err != nil {
		return errors.Wrap(err, "error validating ssh access")
	}
	status.End(ctx, true)

	// Initialize the control plane
	status.Start(ctx, "Configuring control plane node(s) 👀")
	joinCmd, err := kubeadmInit(ctx, cfg)
	if err != nil {
		return err
	}
	if err := kubeApplyNetworking(ctx, cfg); err != nil {
		return err
	}
	if err := kubeConfigGet(ctx, cfg); err != nil {
		return err
	}
	status.End(ctx, true)

	// Initialize the worker node(s)
	status.Start(ctx, "Configuring worker node(s) 🐴")
	for _, node := range cfg.Nodes[1:] {
		if err := kubeadmJoin(ctx, cfg, node, joinCmd); err != nil {
			return err
		}
	}
	status.End(ctx, true)

	// Write the config file to disk.
	configFilePath := config.File(cfg.ID, config.ConfigFileName)
	if err := util.WriteConfig(*cfg, configFilePath, false); err != nil {
		return errors.Wrapf(
			err, "error writing config file to %q", configFilePath)
	}

	tpl := template.Must(template.New("t").Parse(onlineMsgPatt))
	return tpl.Execute(os.Stdout, struct {
		ID         string
		Program    string
		ConfigPath string
	}{
		ID:         cfg.ID,
		ConfigPath: "",
		Program:    os.Args[0],
	})
}

const onlineMsgPatt = `
Access Kubernetes with the following command:
  kubectl --kubeconfig $({{.Program}} config kube {{.ID}})

The nodes may also be accessed with SSH:
  ssh -F $({{.Program}} config ssh {{.ID}}) HOST

Print the available ssh HOST values using:
  {{.Program}} config ssh {{.ID}} --hosts

Finally, the cluster may be deleted with:
  {{.Program}} {{if .ConfigPath}}--config {{.ConfigPath}}{{end}} cluster down {{.ID}}
`
