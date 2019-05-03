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

package lvs_test

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/lvs"
	"vmware.io/sk8/pkg/net/ssh"
	"vmware.io/sk8/pkg/test"
	test_ssh "vmware.io/sk8/pkg/test/ssh"
)

const (
	nicFlagName = "nic"
	defaultNIC  = "eth0"
)

var (
	nicName = flag.String(nicFlagName, defaultNIC, "the name of the public network interface")
)

// ex. go test -v ./pkg/net/lvs -ssh-key "${HOME}/.ssh/id_rsa" -ssh-user "akutz"

func TestMain(m *testing.M) {
	flag.Lookup(test_ssh.AddrFlagName).DefValue = "52.34.10.152"
	flag.Lookup(test_ssh.AddrFlagName).Value.Set("52.34.10.152")
	test.UpdateLogLevel()
	os.Exit(m.Run())
}

func TestCreateAndDeleteRoundRobinTCPService(t *testing.T) {
	if nicName == nil || *nicName == "" {
		t.Fatalf("-%s required", nicFlagName)
	}

	sshCred, err := test_ssh.Credential()
	if err != nil {
		t.Fatal(err)
	}

	client, err := ssh.NewClient(test_ssh.Endpoint(), *sshCred)
	if err != nil {
		t.Fatalf("error getting ssh client for lvs host: %v", err)
	}
	defer client.Close()

	var (
		ctx         = context.Background()
		serviceName = "sk8-1234567-api"
		targets     = []config.ServiceEndpoint{
			{
				Addr: "192.168.3.10",
				Port: 443,
			},
			{
				Addr: "192.168.3.11",
				Port: 443,
			},
		}
		vipAddr = "192.168.2.20"
	)

	port, err := lvs.CreateRoundRobinTCPService(
		ctx, client, *nicName, serviceName, vipAddr)
	if err != nil {
		t.Fatalf("error creating rr tcp service: %v", err)
	}

	t.Logf("service addr %s:%d", vipAddr, port)

	if err := lvs.AddTargetToRoundRobinTCPService(
		ctx, client, serviceName, vipAddr, targets[0]); err != nil {
		t.Errorf("error adding rr service target: %v", err)
		t.Fail()
	}

	if err := lvs.AddTargetToRoundRobinTCPService(
		ctx, client, serviceName, vipAddr, targets[0]); err != nil {
		t.Errorf("error adding rr service target: %v", err)
		t.Fail()
	}

	if err := lvs.AddTargetToRoundRobinTCPService(
		ctx, client, serviceName, vipAddr, targets[1]); err != nil {
		t.Errorf("error adding rr service target: %v", err)
		t.Fail()
	}

	if err := lvs.DeleteTCPService(
		ctx, client, *nicName, serviceName, vipAddr); err != nil {
		t.Fatalf("error deleting rr tcp service: %v", err)
	}
}

func TestGetOrSetTCPService(t *testing.T) {
	if nicName == nil || *nicName == "" {
		t.Fatalf("-%s required", nicFlagName)
	}

	sshCred, err := test_ssh.Credential()
	if err != nil {
		t.Fatal(err)
	}

	client, err := ssh.NewClient(test_ssh.Endpoint(), *sshCred)
	if err != nil {
		t.Fatalf("error getting ssh client for lvs host: %v", err)
	}
	defer client.Close()

	var (
		ctx         = context.Background()
		serviceName = "sk8-1234567-ssh"
		target      = config.ServiceEndpoint{
			Addr: "192.168.3.10",
			Port: 22,
		}
		vipAddr = "192.168.2.20"
	)

	port, err := lvs.CreateRoundRobinTCPService(
		ctx, client, *nicName, serviceName, vipAddr)
	if err != nil {
		t.Fatalf("error creating rr tcp service: %v", err)
	}

	t.Logf("service addr %s:%d", vipAddr, port)

	expObj := &target

	if actObj, err := lvs.SetOrGetTargetTCPService(
		ctx, client, serviceName, vipAddr, target); err != nil {

		t.Errorf("error getting or setting service target: %v", err)
		t.Fail()
	} else if diff := cmp.Diff(expObj, actObj); diff != "" {
		t.Errorf("1st diff\n\n%s\n", diff)
		t.Fail()
	}

	if actObj, err := lvs.SetOrGetTargetTCPService(
		ctx, client, serviceName, vipAddr, target); err != nil {

		t.Errorf("error getting or setting service target: %v", err)
		t.Fail()
	} else if diff := cmp.Diff(expObj, actObj); diff != "" {
		t.Errorf("2nd diff\n\n%s\n", diff)
		t.Fail()
	}

	if err := lvs.DeleteTCPService(
		ctx, client, *nicName, serviceName, vipAddr); err != nil {
		t.Fatalf("error deleting rr tcp service: %v", err)
	}
}
