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
	"strings"

	aws_pkg "github.com/aws/aws-sdk-go/aws"
	aws_elb "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/lvs"
	"vmware.io/sk8/pkg/net/ssh"
	vaws "vmware.io/sk8/pkg/provider/vsphere/aws"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
)

func (a actuator) natDelete(ctx *reqctx) error {
	switch nat := ctx.ccfg.NAT.Object.(type) {
	case *config.LinuxVirtualSwitchConfig:
		if err := a.natDeleteLVS(ctx, nat); err != nil {
			return errors.Wrap(err, "error deleting lvs resources")
		}
	case *vconfig.AWSLoadBalancerConfig:
		if err := a.natDeleteAWS(ctx, nat); err != nil {
			return errors.Wrap(err, "error deleting aws resources")
		}
	}
	return nil
}

func (a actuator) natEnsure(ctx *reqctx) error {
	switch nat := ctx.ccfg.NAT.Object.(type) {
	case *config.LinuxVirtualSwitchConfig:
		return a.natEnsureLVS(ctx, nat)
	case *vconfig.AWSLoadBalancerConfig:
		return a.natEnsureAWS(ctx, nat)
	}
	return nil
}

func (a actuator) natDeleteLVS(
	ctx *reqctx,
	nat *config.LinuxVirtualSwitchConfig) error {

	sshClient, err := ssh.NewClient(
		nat.SSH.SSHEndpoint,
		nat.SSH.SSHCredential)
	if err != nil {
		return errors.Wrap(err, "error dialing lvs host")
	}
	defer sshClient.Close()
	deleteService := func(sid string) error {
		err := lvs.DeleteTCPService(
			ctx,
			sshClient,
			nat.PublicNIC,
			sid,
			nat.PublicIPAddr.String())
		if err != nil {
			return errors.Wrapf(
				err, "error deleting lvs service %q", sid)
		}
		log.WithFields(log.Fields{
			"service": sid,
			"cluster": ctx.cluster.Name,
		}).Info("deleted lvs service")
		return nil
	}
	if err := deleteService(ctx.cluster.Name + "-api"); err != nil {
		return err
	}
	if err := deleteService(ctx.cluster.Name + "-ssh"); err != nil {
		return err
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
		return errors.Wrap(err, "error dialing lvs host")
	}
	defer sshClient.Close()
	createService := func(sid string) (int, error) {
		port, err := lvs.CreateRoundRobinTCPService(
			ctx,
			sshClient,
			nat.PublicNIC,
			sid,
			nat.PublicIPAddr.String())
		if err != nil {
			return 0, errors.Wrapf(
				err, "error creating lvs service %q", sid)
		}
		log.WithFields(log.Fields{
			"service": sid,
			"port":    port,
			"cluster": ctx.cluster.Name,
		}).Info("created lvs service")
		return port, nil
	}
	if len(ctx.cluster.Status.APIEndpoints) == 0 {
		port, err := createService(ctx.cluster.Name + "-api")
		if err != nil {
			return err
		}
		ctx.cluster.Status.APIEndpoints = append(
			ctx.cluster.Status.APIEndpoints,
			capi.APIEndpoint{
				Host: nat.SSH.Addr,
				Port: port,
			})
	}
	if ctx.csta.SSH == nil {
		port, err := createService(ctx.cluster.Name + "-ssh")
		if err != nil {
			return err
		}
		ctx.csta.SSH = &config.SSHEndpoint{
			ServiceEndpoint: config.ServiceEndpoint{
				Addr: nat.SSH.Addr,
				Port: int32(port),
			},
		}
	}

	return nil
}

func (a actuator) awsEnsureSession(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) error {

	ses, err := vaws.NewSession(ctx, nat)
	if err != nil {
		return err
	}
	ctx.csta.AWS = &vconfig.AWSLoadBalancerStatus{
		Session: ses,
		CanOwn:  make(chan struct{}, 1),
		Online:  make(chan struct{}),
	}
	ctx.csta.AWS.CanOwn <- struct{}{}
	return nil
}

func (a actuator) natDeleteAWS(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) error {

	if err := a.awsEnsureSession(ctx, nat); err != nil {
		return err
	}
	if err := a.awsDestroyLoadBalancer(ctx, nat); err != nil {
		return errors.Wrap(err, "error destroying load balancer")
	}
	if err := a.awsDestroyTargetGroups(ctx, nat); err != nil {
		return errors.Wrap(err, "error destroying target groups")
	}
	return nil
}

func (a actuator) natEnsureAWS(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) error {

	if err := a.awsEnsureSession(ctx, nat); err != nil {
		return err
	}
	awsEnsureListeners, err := a.awsEnsureLoadBalancer(ctx, nat)
	if err != nil {
		return err
	}
	if err := a.awsEnsureTargetGroups(ctx, nat); err != nil {
		return err
	}
	if err := awsEnsureListeners(); err != nil {
		return err
	}
	if len(ctx.cluster.Status.APIEndpoints) == 0 {
		ctx.cluster.Status.APIEndpoints = append(
			ctx.cluster.Status.APIEndpoints,
			capi.APIEndpoint{
				Host: *ctx.csta.AWS.LoadBalancer.DNSName,
				Port: 443,
			})
	}
	if ctx.csta.SSH == nil {
		ctx.csta.SSH = &config.SSHEndpoint{
			ServiceEndpoint: config.ServiceEndpoint{
				Addr: *ctx.csta.AWS.LoadBalancer.DNSName,
				Port: 22,
			},
		}
	}
	return nil
}

func (a actuator) awsEnsureLoadBalancer(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) (func() error, error) {

	ensureLis := func() error { return nil }

	lb, err := vaws.FindLoadBalancer(
		ctx, ctx.csta.AWS.Session, nat, ctx.cluster.Name)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return ensureLis, err
		}
	}
	if lb != nil {
		ctx.csta.AWS.LoadBalancer = lb
		return ensureLis, nil
	}
	if err := a.awsCreateLoadBalancer(ctx, nat); err != nil {
		return ensureLis, err
	}

	createListener := func(tg *aws_elb.TargetGroup) error {
		if err := a.awsCreateListener(ctx, nat, tg); err != nil {
			return errors.Wrapf(
				err,
				"error creating listener for lb=%q tg=%q",
				*ctx.csta.AWS.LoadBalancer.LoadBalancerName,
				*tg.TargetGroupName)
		}
		return nil
	}

	ensureLis = func() error {
		if err := createListener(ctx.csta.AWS.API); err != nil {
			return err
		}
		if err := createListener(ctx.csta.AWS.SSH); err != nil {
			return err
		}
		return nil
	}

	return ensureLis, nil
}

func (a actuator) awsEnsureTargetGroups(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) error {

	// Define the port used to issue health checks.
	var healthCheckPort *string
	if p := nat.HealthCheckPort; p > 0 {
		healthCheckPort = aws_pkg.String(fmt.Sprintf("%d", p))
	}

	tgAPI, tgSSH, err := vaws.FindTargetGroups(
		ctx, ctx.csta.AWS.Session, nat,
		ctx.cluster.Name,
		ctx.csta.AWS.LoadBalancer)
	if err != nil {
		//if !strings.Contains(err.Error(), "not found") {
		//	return err
		//}
		return err
	}

	if tgAPI != nil {
		ctx.csta.AWS.API = tgAPI
	} else if err := a.awsCreateTargetGroup(
		ctx, nat, healthCheckPort, "api", 443); err != nil {
		return err
	}

	if tgSSH != nil {
		ctx.csta.AWS.SSH = tgSSH
	} else if err := a.awsCreateTargetGroup(
		ctx, nat, healthCheckPort, "ssh", 22); err != nil {
		return err
	}

	return nil
}

func (a actuator) awsCreateLoadBalancer(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) error {

	elb := aws_elb.New(ctx.csta.AWS.Session)
	input := &aws_elb.CreateLoadBalancerInput{
		Name:          aws_pkg.String(ctx.cluster.Name),
		IpAddressType: aws_pkg.String(aws_elb.IpAddressTypeIpv4),
		Scheme:        aws_pkg.String(aws_elb.LoadBalancerSchemeEnumInternetFacing),
		Type:          aws_pkg.String(aws_elb.LoadBalancerTypeEnumNetwork),
		Subnets:       []*string{aws_pkg.String(nat.SubnetID)},
	}
	log.WithField("input", input).Debug("creating aws load balancer")
	output, err := elb.CreateLoadBalancerWithContext(ctx, input)
	if err != nil {
		return errors.Wrap(err, "error creating load balancer")
	}
	if len(output.LoadBalancers) == 0 {
		return errors.New("create load balancer returned zero results")
	}
	ctx.csta.AWS.LoadBalancer = output.LoadBalancers[0]
	return nil
}

func (a actuator) awsCreateTargetGroup(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig,
	healthCheckPort *string,
	suffix string, port int64) error {

	elb := aws_elb.New(ctx.csta.AWS.Session)
	name := fmt.Sprintf("%s-%s", ctx.cluster.Name, suffix)
	input := &aws_elb.CreateTargetGroupInput{
		Name:                aws_pkg.String(name),
		HealthCheckPort:     healthCheckPort,
		HealthCheckEnabled:  aws_pkg.Bool(true),
		HealthCheckProtocol: aws_pkg.String(aws_elb.ProtocolEnumTcp),
		Protocol:            aws_pkg.String(aws_elb.ProtocolEnumTcp),
		Port:                aws_pkg.Int64(port),
		TargetType:          aws_pkg.String(aws_elb.TargetTypeEnumIp),
		VpcId:               aws_pkg.String(nat.VpcID),
	}
	log.WithField("input", input).Debug("creating aws target group")
	output, err := elb.CreateTargetGroupWithContext(ctx, input)
	if err != nil {
		return errors.Wrapf(err, "error creating target group %q", name)
	}
	if len(output.TargetGroups) == 0 {
		return errors.Errorf(
			"create target group %q returned zero results", name)
	}

	switch suffix {
	case "api":
		ctx.csta.AWS.API = output.TargetGroups[0]
	case "ssh":
		ctx.csta.AWS.SSH = output.TargetGroups[0]
	}

	return nil
}

func (a actuator) awsCreateListener(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig,
	tg *aws_elb.TargetGroup) error {

	elb := aws_elb.New(ctx.csta.AWS.Session)
	input := &aws_elb.CreateListenerInput{
		Port:            tg.Port,
		Protocol:        tg.Protocol,
		LoadBalancerArn: ctx.csta.AWS.LoadBalancer.LoadBalancerArn,
		DefaultActions: []*aws_elb.Action{
			{
				Type:           aws_pkg.String(aws_elb.ActionTypeEnumForward),
				TargetGroupArn: tg.TargetGroupArn,
			},
		},
	}
	output, err := elb.CreateListenerWithContext(ctx, input)
	if err != nil {
		return errors.Wrapf(
			err, "error creating listener for %q",
			*ctx.csta.AWS.LoadBalancer.LoadBalancerName)
	}
	if len(output.Listeners) == 0 {
		return errors.Errorf(
			"create listener for %q returned zero results",
			*ctx.csta.AWS.LoadBalancer.LoadBalancerName)
	}
	return nil
}

func (a actuator) awsDestroyLoadBalancer(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) error {

	if lb, err := vaws.FindLoadBalancer(
		ctx, ctx.csta.AWS.Session, nat, ctx.cluster.Name); err == nil {
		elb := aws_elb.New(ctx.csta.AWS.Session)
		if _, err := elb.DeleteLoadBalancerWithContext(
			ctx, &aws_elb.DeleteLoadBalancerInput{
				LoadBalancerArn: lb.LoadBalancerArn,
			}); err != nil {
			return errors.Wrapf(
				err,
				"error deleting load balancer %q",
				*lb.LoadBalancerName)
		}
	}
	return nil
}

func (a actuator) awsDestroyTargetGroups(
	ctx *reqctx,
	nat *vconfig.AWSLoadBalancerConfig) error {

	elb := aws_elb.New(ctx.csta.AWS.Session)
	input := &aws_elb.DescribeTargetGroupsInput{
		Names: []*string{
			aws_pkg.String(fmt.Sprintf("%s-api", ctx.cluster.Name)),
			aws_pkg.String(fmt.Sprintf("%s-ssh", ctx.cluster.Name)),
		},
	}
	output, _ := elb.DescribeTargetGroupsWithContext(ctx, input)
	if output != nil {
		for _, tg := range output.TargetGroups {
			if _, err := elb.DeleteTargetGroupWithContext(
				ctx, &aws_elb.DeleteTargetGroupInput{
					TargetGroupArn: tg.TargetGroupArn,
				}); err != nil {
				return errors.Wrapf(
					err,
					"error deleting target group %q",
					*tg.TargetGroupName)
			}
		}
	}
	return nil
}
