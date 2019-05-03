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
	"fmt"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"

	"vmware.io/sk8/pkg/config"
)

func (a actuator) sshEnsure(ctx *reqctx) error {
	configDir := ctx.cluster.Labels[config.ConfigDirLabelName]
	if configDir == "" {
		return nil
	}

	// Write the SSH config to disk.
	cfgPath := path.Join(configDir, "ssh.conf")
	knownHostsPath := path.Join(configDir, "ssh.known_hosts")
	if err := ioutil.WriteFile(
		cfgPath,
		[]byte(fmt.Sprintf(sshConfigHeaderFormat, knownHostsPath)),
		0640); err != nil {
		return errors.Wrapf(
			err,
			"error writing ssh config to %q",
			cfgPath)
	}

	// Write the public and private keys to disk.
	keyPath := path.Join(configDir, "ssh.key")
	if err := ioutil.WriteFile(
		keyPath,
		ctx.ccfg.SSH.PrivateKey,
		0400); err != nil {
		return errors.Wrapf(
			err,
			"error writing ssh private key file to %q",
			keyPath)
	}
	pubPath := path.Join(configDir, "ssh.pub")
	if err := ioutil.WriteFile(
		pubPath,
		ctx.ccfg.SSH.PublicKey,
		0440); err != nil {
		return errors.Wrapf(
			err,
			"error writing ssh public key file to %q",
			pubPath)
	}

	return nil
}

const sshConfigHeaderFormat = `ServerAliveInterval 300
TCPKeepAlive        no
UserKnownHostsFile  %s

`
