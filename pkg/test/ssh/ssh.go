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

package ssh

import (
	"flag"
	"io/ioutil"
	"os/user"

	"github.com/pkg/errors"

	"vmware.io/sk8/pkg/config"
)

const (
	// AddrFlagName is the name of the SSH address flag.
	AddrFlagName = "ssh-addr"
	defaultPort  = 22
	keyFlagName  = "ssh-key"
	portFlagName = "ssh-port"
	userFlagName = "ssh-user"
)

var (
	addr     = flag.String(AddrFlagName, "50.112.88.129", "the address of the ssh server")
	keyFile  = flag.String(keyFlagName, "", "path to an ssh private key file")
	port     = flag.Int(portFlagName, defaultPort, "the port on which the ssh server is listening")
	username *string
)

func init() {
	var defaultUsername string
	if u, err := user.Current(); err == nil {
		defaultUsername = u.Username
	}
	username = flag.String(userFlagName, defaultUsername, "the username to use for ssh")
}

// Credential returns the SSHCredential loaded from the command lind args.
func Credential() (*config.SSHCredential, error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if keyFile == nil || *keyFile == "" {
		return nil, errors.Errorf("-%s required", keyFlagName)
	}
	if username == nil || *username == "" {
		return nil, errors.Errorf("-%s required", userFlagName)
	}

	buf, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error reading ssh key file %q", *keyFile)
	}

	return &config.SSHCredential{
		PrivateKey: buf,
		Username:   *username,
	}, nil
}

// Endpoint returns a common SSHEndpoint used with net.ssh and net.lvs tests.
func Endpoint() config.SSHEndpoint {
	if !flag.Parsed() {
		flag.Parse()
	}
	return config.SSHEndpoint{
		ServiceEndpoint: config.ServiceEndpoint{
			Addr: *addr,
			Port: int32(*port),
		},
	}
}
