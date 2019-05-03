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
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/ssh"
)

func (a actuator) sshEnsure(ctx *reqctx) error {
	if err := a.sshEnsureOnline(ctx); err != nil {
		return err
	}
	if err := a.sshEnsureRemoteKey(ctx); err != nil {
		return err
	}
	return a.sshEnsureLocalConf(ctx)
}

func (a actuator) sshEnsureOnline(ctx *reqctx) error {
	sleep := 5 * time.Second
	done := make(chan struct{})
	go func() {
		for ctx.Err() == nil {
			sshClient, err := ssh.NewClient(*ctx.msta.SSH, ctx.ccfg.SSH)
			if err != nil {
				log.WithError(err).Debug("ssh not ready")
				time.Sleep(sleep)
			} else {
				sshClient.Close()
				close(done)
				return
			}
		}
	}()
	for {
		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (a actuator) sshEnsureRemoteKey(ctx *reqctx) error {
	keyPath := fmt.Sprintf("/home/%s/.ssh/id_rsa", ctx.ccfg.SSH.Username)
	keyMode := os.FileMode(0400)

	sshClient, err := ssh.NewClient(*ctx.msta.SSH, ctx.ccfg.SSH)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	if err := ssh.Upload(
		ctx, sshClient,
		ctx.ccfg.SSH.PrivateKey, keyPath,
		ctx.ccfg.SSH.Username, ctx.ccfg.SSH.Username,
		keyMode); err != nil {
		return errors.Wrapf(
			err, "error uploading key to %q", ctx.machine.Name)
	}

	return nil
}

func (a actuator) sshEnsureLocalConf(ctx *reqctx) error {
	configDir := ctx.cluster.Labels[config.ConfigDirLabelName]
	if configDir == "" {
		return nil
	}

	ctx.csta.SSHConfigMu.Lock()
	defer ctx.csta.SSHConfigMu.Unlock()

	sshConfPath := path.Join(configDir, "ssh.conf")
	sshKeyPath := path.Join(configDir, "ssh.key")
	f, err := os.OpenFile(
		sshConfPath,
		os.O_RDWR|os.O_APPEND,
		0640)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f,
		sshConfigEntryFormat,
		strings.Split(ctx.machine.Name, ".")[0],
		ctx.msta.SSH.Addr,
		ctx.msta.SSH.Port,
		ctx.ccfg.SSH.Username,
		sshKeyPath)

	if ctx.msta.SSH.ProxyHost == nil {
		fmt.Fprintf(f,
			sshConfigEntryFormat,
			"proxy",
			ctx.msta.SSH.Addr,
			ctx.msta.SSH.Port,
			ctx.ccfg.SSH.Username,
			sshKeyPath)
	} else {
		fmt.Fprintf(f, sshConfigProxyFormat, sshConfPath)
	}

	return nil
}

const sshConfigEntryFormat = `host %s
  HostName     %s
  Port         %d
  User         %s
  IdentityFile %s
  PasswordAuthentication no
`
const sshConfigProxyFormat = `  ProxyCommand  ssh -F %q -q -W %%h:%%p proxy
`
