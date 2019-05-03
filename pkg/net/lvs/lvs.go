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

package lvs

import (
	"bytes"
	"context"
	"net"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/ssh"
)

// CreateRoundRobinTCPService creates a new round-robin service for the given
// targets and returns the port on which the new service listens.
//
// The serviceStartPort value is the lowest port the service will use.
func CreateRoundRobinTCPService(
	ctx context.Context,
	client *ssh.Client,
	publicNIC, serviceID, vipAddr string) (int, error) {

	tpl := template.Must(template.New("t").Parse(createRRTCPServiceCmdPatt))
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, struct {
		NIC string
		SID string
		VIP string
	}{
		NIC: publicNIC,
		SID: serviceID,
		VIP: vipAddr,
	}); err != nil {
		return 0, errors.Wrap(err, "error parsing create lvs service template")
	}

	stdout := &bytes.Buffer{}
	if err := ssh.Run(ctx, client, nil, stdout, nil, buf.String()); err != nil {
		return 0, errors.Wrap(err, "error creating lvs service")
	}

	szPort := strings.TrimSpace(stdout.String())
	port, err := strconv.Atoi(szPort)
	if err != nil {
		return 0, errors.Wrapf(
			err, "error capturing lvs service port %q", szPort)
	}

	return port, nil
}

// AddTargetToRoundRobinTCPService adds a target to an existing round-robin
// service.
func AddTargetToRoundRobinTCPService(
	ctx context.Context,
	client *ssh.Client,
	serviceID, vipAddr string,
	target config.ServiceEndpoint) error {

	tpl := template.Must(template.New("t").Parse(addRRTCPServiceCmdPatt))
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, struct {
		SID    string
		VIP    string
		Target config.ServiceEndpoint
	}{
		SID:    serviceID,
		VIP:    vipAddr,
		Target: target,
	}); err != nil {
		return errors.Wrap(
			err, "error parsing add lvs service target template")
	}

	if err := ssh.Run(ctx, client, nil, nil, nil, buf.String()); err != nil {
		return errors.Wrap(err, "error adding lvs service target")
	}

	return nil
}

// SetOrGetTargetTCPService sets a single target on an LVS service or returns
// the existing target.
func SetOrGetTargetTCPService(
	ctx context.Context,
	client *ssh.Client,
	serviceID, vipAddr string,
	target config.ServiceEndpoint) (*config.ServiceEndpoint, error) {

	tpl := template.Must(template.New("t").Parse(setOrGetTCPServiceCmdPatt))
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, struct {
		SID    string
		VIP    string
		Target config.ServiceEndpoint
	}{
		SID:    serviceID,
		VIP:    vipAddr,
		Target: target,
	}); err != nil {
		return nil, errors.Wrap(
			err, "error parsing setOrGet lvs service target template")
	}

	stdout := &bytes.Buffer{}
	if err := ssh.Run(ctx, client, nil, stdout, nil, buf.String()); err != nil {
		return nil, errors.Wrap(
			err, "error setting or getting single lvs service target")
	}

	addrAndPort := strings.TrimSpace(stdout.String())
	addr, szPort, err := net.SplitHostPort(addrAndPort)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error splitting host and port %q", addrAndPort)
	}

	port, err := strconv.Atoi(szPort)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error parsing port as number %q", szPort)
	}

	return &config.ServiceEndpoint{
		Addr: addr,
		Port: int32(port),
	}, nil
}

// DeleteTCPService deletes the service with the specified port and all of the
// associated, real servers.
func DeleteTCPService(
	ctx context.Context,
	client *ssh.Client,
	publicNIC, serviceID, vipAddr string) error {

	tpl := template.Must(template.New("t").Parse(deleteRRTCPServiceCmdPatt))
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, struct {
		NIC string
		SID string
		VIP string
	}{
		NIC: publicNIC,
		SID: serviceID,
		VIP: vipAddr,
	}); err != nil {
		return errors.Wrap(err, "error parsing delete lvs service template")
	}

	if err := ssh.Run(
		ctx, client, nil, nil, nil,
		buf.String()); err != nil {
		return errors.Wrap(err, "error deleting ipvs service")
	}

	return nil
}
