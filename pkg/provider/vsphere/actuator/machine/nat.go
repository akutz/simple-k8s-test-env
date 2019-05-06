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
	aws_pkg "github.com/aws/aws-sdk-go/aws"
	aws_elb "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/lvs"
	"vmware.io/sk8/pkg/net/ssh"
	vaws "vmware.io/sk8/pkg/provider/vsphere/aws"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
	"vmware.io/sk8/pkg/status"
)

func (a actuator) natEnsure(ctx *reqctx) error {
	// Configure the NAT provider.
	switch nat := ctx.ccfg.NAT.Object.(type) {
	case *config.LinuxVirtualSwitchConfig:
		if err := a.natEnsureLVS(ctx, nat); err != nil {
			return errors.Wrap(err, "error ensuring nat lvs")
		}
	case *vconfig.AWSLoadBalancerConfig:
		if err := a.natEnsureAWS(ctx, nat); err != nil {
			return errors.Wrap(err, "error ensuring nat aws")
		}
	default:
		ctx.msta.SSH = &config.SSHEndpoint{
			ServiceEndpoint: config.ServiceEndpoint{
				Addr: ctx.machine.Status.Addresses[0].Address,
				Port: 22,
			},
		}
	}

	return nil
}

func (a actuator) natEnsureLVS(
	ctx *reqctx,
	nat *config.LinuxVirtualSwitchConfig) error {

	sshClient, err := ssh.NewClient(
		nat.SSH.SSHEndpoint,
		nat.SSH.SSHCredential)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	// Get the LVS target interface for this machine.
	lvsTgt := config.ServiceEndpoint{
		Port: 22,
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
	sshTgt, err := lvs.SetOrGetTargetTCPService(
		ctx,
		sshClient,
		ctx.cluster.Name+"-ssh",
		nat.PublicIPAddr.String(),
		lvsTgt)
	if err != nil {
		return err
	}

	// If the returned SSH target is *this* machine then the
	// machine's SSH endpoint config does not require a bastion
	// proxy. Otherwise it does.
	if *sshTgt == lvsTgt {
		ctx.msta.SSH = ctx.csta.SSH
	} else {
		ctx.msta.SSH = &config.SSHEndpoint{
			ServiceEndpoint: lvsTgt,
			ProxyHost:       ctx.csta.SSH,
		}
	}

	return nil
}

func (a actuator) natEnsureAWS(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) error {

	if err := a.awsWaitLoadBalancer(ctx, nat); err != nil {
		return err
	}

	// Get the IP address for this machine.
	var ipAddr string
	for _, addr := range ctx.machine.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			ipAddr = addr.Address
		}
	}
	if ipAddr == "" {
		return errors.Errorf("no IP address for %q", ctx.machine.Name)
	}

	elb := aws_elb.New(ctx.csta.AWS.Session)
	registerTarget := func(tg *aws_elb.TargetGroup) error {
		input := &aws_elb.RegisterTargetsInput{
			TargetGroupArn: tg.TargetGroupArn,
			Targets: []*aws_elb.TargetDescription{
				{
					Id:               aws_pkg.String(ipAddr),
					Port:             tg.Port,
					AvailabilityZone: ctx.csta.AWS.LoadBalancer.AvailabilityZones[0].ZoneName,
				},
			},
		}
		log.WithField("input", input).Debug("registering aws target")
		if _, err := elb.RegisterTargetsWithContext(
			ctx, input); err != nil {
			return errors.Wrapf(
				err, "error registering machine %q for target group %q",
				*tg.TargetGroupName, ctx.machine.Name)
		}
		return nil
	}

	if ctx.role.Has(config.MachineRoleControlPlane) {
		ok, err := a.awsIsTargetRegistered(ctx, nat, ctx.csta.AWS.API, ipAddr)
		if err != nil {
			return err
		}
		if !ok {
			if err := registerTarget(ctx.csta.AWS.API); err != nil {
				return err
			}
		}
	}

	// Only one machine needs to be the SSH target for this cluster. Check
	// first to see if an SSH target is registered, and if not, register
	// this machine as that target.
	ctx.csta.SSHConfigMu.Lock()
	defer ctx.csta.SSHConfigMu.Unlock()

	// Get the list of registered targets for the SSH target group.
	sshTargets, err := a.awsGetRegisteredTargets(ctx, nat, ctx.csta.AWS.SSH)
	if err != nil {
		return err
	}

	log.WithField("sshTargets", sshTargets).Debug(
		"got registered aws SSH targets")

	// If there are *no* registered SSH targets then register this machine
	// as the SSH target group's target.
	if len(sshTargets) == 0 {
		if err := registerTarget(ctx.csta.AWS.SSH); err != nil {
			return err
		}
		ctx.msta.SSH = ctx.csta.SSH
		return nil
	}

	// Check to see if this machine is already registered as the SSH target.
	// If so, the machine's SSH endpoint is the cluster SSH endpoint.
	for _, sshTgt := range sshTargets {
		if ipAddr == sshTgt {
			ctx.msta.SSH = ctx.csta.SSH
			return nil
		}
	}

	// The machine is not the SSH endpoint, so make its endpoint use the
	// SSH target as the bastion host.
	ctx.msta.SSH = &config.SSHEndpoint{
		ServiceEndpoint: config.ServiceEndpoint{
			Addr: ipAddr,
			Port: 22,
		},
		ProxyHost: ctx.csta.SSH,
	}

	return nil

}

func (a actuator) awsIsTargetRegistered(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig,
	tg *aws_elb.TargetGroup,
	ipAddr string) (bool, error) {

	targetIDs, err := a.awsGetRegisteredTargets(ctx, nat, tg)
	if err != nil {
		return false, err
	}
	for _, tid := range targetIDs {
		if ipAddr == tid {
			return true, nil
		}
	}
	return false, nil
}

func (a actuator) awsGetRegisteredTargets(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig,
	tg *aws_elb.TargetGroup) ([]string, error) {

	elb := aws_elb.New(ctx.csta.AWS.Session)
	input := &aws_elb.DescribeTargetHealthInput{
		TargetGroupArn: tg.TargetGroupArn,
	}
	output, err := elb.DescribeTargetHealthWithContext(ctx, input)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error describing target health for %q", *tg.TargetGroupName)
	}
	targetIDs := make([]string, len(output.TargetHealthDescriptions))
	for i, thd := range output.TargetHealthDescriptions {
		if t := thd.Target; t != nil {
			if id := t.Id; id != nil {
				targetIDs[i] = *id
			} else {
				targetIDs[i] = ""
			}
		}
	}
	return targetIDs, nil
}

func (a actuator) awsWaitLoadBalancer(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) error {

	// Only one machine needs to wait until the load balancer is online.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ctx.csta.AWS.CanOwn:
		defer status.End(ctx, false)
		status.End(ctx, true)
		status.Start(ctx, "Wait until AWS load balancer is ready 📡")
		log.Debug("waiting until the aws load balancer is ready")
		if err := vaws.WaitLoadBalancer(
			ctx, ctx.csta.AWS.Session,
			ctx.csta.AWS.LoadBalancer); err != nil {
			return err
		}
		close(ctx.csta.AWS.Online)
		status.End(ctx, true)
	case <-ctx.csta.AWS.Online:
	}

	return nil
}
