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

package ssh_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/pkg/errors"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/ssh"
	"vmware.io/sk8/pkg/test"
	test_ssh "vmware.io/sk8/pkg/test/ssh"
)

const cmd = `sh -c 'printf "%%s@%%s\n" "$(whoami)" "$(hostname -s)"'`

// ex. go test -v ./pkg/net/ssh -ssh-key "${HOME}/.ssh/id_rsa" -ssh-user "akutz"

func TestMain(m *testing.M) {
	test.UpdateLogLevel()
	os.Exit(m.Run())
}

func TestDial(t *testing.T) {
	sshCred, err := test_ssh.Credential()
	if err != nil {
		t.Fatal(err)
	}
	sshEndpoint := test_ssh.Endpoint()

	cases := []struct {
		TestName    string
		SSH         config.SSHEndpoint
		ExpectError bool
	}{
		{
			TestName: "Direct",
			SSH:      sshEndpoint,
		},
		{
			TestName: "Bastion",
			SSH: config.SSHEndpoint{
				ServiceEndpoint: config.ServiceEndpoint{
					Addr: "192.168.2.3",
					Port: 22,
				},
				ProxyHost: &sshEndpoint,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.TestName, func(t *testing.T) {
			handleTestCase := func() error {
				client, err := ssh.NewClient(c.SSH, *sshCred)
				if err != nil {
					return errors.Wrapf(
						err, "error getting ssh client for %q", c.SSH.Addr)
				}
				defer client.Close()
				stdout := &bytes.Buffer{}
				if err := ssh.Run(
					context.Background(), client,
					nil, stdout, nil, cmd); err != nil {
					return err
				}
				t.Log(stdout.String())
				return nil
			}
			if err := handleTestCase(); err != nil && !c.ExpectError {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && c.ExpectError {
				t.Fatalf("unexpected lack or error")
			}
		})
	}
}
