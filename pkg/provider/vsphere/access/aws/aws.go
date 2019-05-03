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

package access

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	aws_pkg "github.com/aws/aws-sdk-go/aws"
	aws_credentials "github.com/aws/aws-sdk-go/aws/credentials"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_elb "github.com/aws/aws-sdk-go/service/elbv2"

	"vmware.io/sk8/pkg/config"
)

type awsProvider struct {
	cfg *config.Cluster
	ses *aws_session.Session

	lb     *aws_elb.LoadBalancer
	apiLis *aws_elb.Listener
	sshLis *aws_elb.Listener
	apiGrp *aws_elb.TargetGroup
	sshGrp *aws_elb.TargetGroup
	secGrp *string
}

func newAWSProvider(
	ctx context.Context,
	cfg *config.Cluster,
	createMissing bool) (Provider, error) {

	p := &awsProvider{cfg: cfg}

	// Create a new session
	ses, err := aws_session.NewSession(&aws_pkg.Config{
		Region:     aws_pkg.String(cfg.AWS.Region),
		MaxRetries: aws_pkg.Int(cfg.AWS.MaxRetries),
		Credentials: aws_credentials.NewChainCredentials(
			[]aws_credentials.Provider{
				&aws_credentials.StaticProvider{
					Value: aws_credentials.Value{
						AccessKeyID:     p.cfg.AWS.AccessKeyID,
						SecretAccessKey: p.cfg.AWS.SecretAccessKey,
					},
				},
				&aws_credentials.EnvProvider{},
				&aws_credentials.SharedCredentialsProvider{},
			},
		),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating AWS session")
	}
	p.ses = ses

	if createMissing {
		return p, p.ensureLoadBalancer(ctx)
	}

	return p, nil
}

func (p *awsProvider) Up(ctx context.Context) error {
	return p.ensureTargets(ctx)
}

func (p *awsProvider) Down(ctx context.Context) error {
	err := p.destroyLoadBalancer(ctx)
	if err := p.destoryTargetGroups(ctx); err != nil {
		return err
	}
	return err
}

func (p *awsProvider) Wait(ctx context.Context) error {
	return p.waitUntilOnline(ctx)
}

func (p *awsProvider) String() string {
	return *p.lb.LoadBalancerName
}

func (p *awsProvider) ensureLoadBalancer(ctx context.Context) error {
	ok, err := p.findLoadBalancer(ctx)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return p.createLoadBalancer(ctx)
}

func (p *awsProvider) ensureTargets(ctx context.Context) error {
	if err := p.registerTargetsAPI(ctx); err != nil {
		return err
	}
	if err := p.registerTargetsSSH(ctx); err != nil {
		return err
	}
	return nil
}

func (p *awsProvider) findLoadBalancer(ctx context.Context) (bool, error) {
	elb := aws_elb.New(p.ses)

	input := &aws_elb.DescribeLoadBalancersInput{
		Names: []*string{aws_pkg.String(p.cfg.ID)},
	}
	output, err := elb.DescribeLoadBalancersWithContext(ctx, input)
	if err != nil {
		return false, errors.Wrap(err, "error finding existing load balancer")
	}
	if len(output.LoadBalancers) > 0 {
		p.lb = output.LoadBalancers[0]
		return true, nil
	}
	return false, nil
}

func (p *awsProvider) createLoadBalancer(ctx context.Context) error {
	elb := aws_elb.New(p.ses)
	input := &aws_elb.CreateLoadBalancerInput{
		Name:          aws_pkg.String(p.cfg.ID),
		IpAddressType: aws_pkg.String(aws_elb.IpAddressTypeIpv4),
		Scheme:        aws_pkg.String(aws_elb.LoadBalancerSchemeEnumInternetFacing),
		Type:          aws_pkg.String(aws_elb.LoadBalancerTypeEnumNetwork),
		Subnets:       []*string{aws_pkg.String(p.cfg.AWS.SubnetID)},
	}
	output, err := elb.CreateLoadBalancerWithContext(ctx, input)
	if err != nil {
		return errors.Wrap(err, "error creating load balancer")
	}
	if len(output.LoadBalancers) == 0 {
		return errors.New("create load balancer returned zero results")
	}
	p.lb = output.LoadBalancers[0]

	// Define the port used to issue health checks.
	var healthCheckPort *string
	if p := p.cfg.HealthCheckPort; p > 0 {
		healthCheckPort = aws_pkg.String(fmt.Sprintf("%d", p))
	}

	// Create the API server target group.
	{
		input := &aws_elb.CreateTargetGroupInput{
			Name:                aws_pkg.String(fmt.Sprintf("%s-api", p.cfg.ID)),
			HealthCheckPort:     healthCheckPort,
			HealthCheckEnabled:  aws_pkg.Bool(true),
			HealthCheckProtocol: aws_pkg.String(aws_elb.ProtocolEnumTcp),
			Protocol:            aws_pkg.String(aws_elb.ProtocolEnumTcp),
			Port:                aws_pkg.Int64(443),
			TargetType:          aws_pkg.String(aws_elb.TargetTypeEnumIp),
			VpcId:               aws_pkg.String(p.cfg.AWS.VpcID),
		}
		output, err := elb.CreateTargetGroupWithContext(ctx, input)
		if err != nil {
			return errors.Wrap(err, "error creating API target group")
		}
		if len(output.TargetGroups) == 0 {
			return errors.New("create API target group returned zero results")
		}
		p.apiGrp = output.TargetGroups[0]
	}

	// Create the API server listener.
	{
		input := &aws_elb.CreateListenerInput{
			Port:            p.apiGrp.Port,
			Protocol:        p.apiGrp.Protocol,
			LoadBalancerArn: p.lb.LoadBalancerArn,
			DefaultActions: []*aws_elb.Action{
				{
					Type:           aws_pkg.String(aws_elb.ActionTypeEnumForward),
					TargetGroupArn: p.apiGrp.TargetGroupArn,
				},
			},
		}
		output, err := elb.CreateListenerWithContext(ctx, input)
		if err != nil {
			return errors.Wrap(err, "error creating API listener")
		}
		if len(output.Listeners) == 0 {
			return errors.New("create API listener returned zero results")
		}
		p.apiLis = output.Listeners[0]
	}

	// Create the SSH target groups.
	{
		input := &aws_elb.CreateTargetGroupInput{
			Name:                aws_pkg.String(fmt.Sprintf("%s-ssh", p.cfg.ID)),
			HealthCheckPort:     healthCheckPort,
			HealthCheckEnabled:  aws_pkg.Bool(true),
			HealthCheckProtocol: aws_pkg.String(aws_elb.ProtocolEnumTcp),
			Protocol:            aws_pkg.String(aws_elb.ProtocolEnumTcp),
			Port:                aws_pkg.Int64(22),
			TargetType:          aws_pkg.String(aws_elb.TargetTypeEnumIp),
			VpcId:               aws_pkg.String(p.cfg.AWS.VpcID),
		}
		output, err := elb.CreateTargetGroupWithContext(ctx, input)
		if err != nil {
			return errors.Wrap(err, "error creating SSH target group")
		}
		if len(output.TargetGroups) == 0 {
			return errors.New("create SSH target group returned zero results")
		}
		p.sshGrp = output.TargetGroups[0]
	}

	// Create the SSH server listener.
	{
		input := &aws_elb.CreateListenerInput{
			Port:            p.sshGrp.Port,
			Protocol:        p.sshGrp.Protocol,
			LoadBalancerArn: p.lb.LoadBalancerArn,
			DefaultActions: []*aws_elb.Action{
				{
					Type:           aws_pkg.String(aws_elb.ActionTypeEnumForward),
					TargetGroupArn: p.sshGrp.TargetGroupArn,
				},
			},
		}
		output, err := elb.CreateListenerWithContext(ctx, input)
		if err != nil {
			return errors.Wrap(err, "error creating SSH listener")
		}
		if len(output.Listeners) == 0 {
			return errors.New("create SSH listener returned zero results")
		}
		p.sshLis = output.Listeners[0]
	}

	return nil
}

func (p *awsProvider) waitUntilOnline(ctx context.Context) error {
	elb := aws_elb.New(p.ses)
	done := make(chan struct{})
	errs := make(chan error)
	input := &aws_elb.DescribeLoadBalancersInput{
		LoadBalancerArns: []*string{p.lb.LoadBalancerArn},
	}
	go func() {
		for ctx.Err() == nil {
			output, err := elb.DescribeLoadBalancersWithContext(ctx, input)
			if err != nil {
				errs <- errors.Wrap(err, "error describing load balancers")
				return
			}
			if len(output.LoadBalancers) == 0 ||
				output.LoadBalancers[0].State == nil ||
				output.LoadBalancers[0].State.Code == nil {
				time.Sleep(5 * time.Second)
				continue
			}

			code := *output.LoadBalancers[0].State.Code
			log.WithField("code", code).Debug("load balancer state")

			switch code {
			case aws_elb.LoadBalancerStateEnumActive:
				close(done)
				return
			case aws_elb.LoadBalancerStateEnumProvisioning:
				time.Sleep(5 * time.Second)
			default:
				errs <- errors.Errorf(
					"load balancer unexpected state %s", code)
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return nil
		case err := <-errs:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *awsProvider) registerTargetsAPI(ctx context.Context) error {

	targets := []*aws_elb.TargetDescription{}
	for _, node := range p.cfg.Nodes {
		switch node.Role {
		case config.BothRole, config.ControlPlaneRole:
			targets = append(targets, &aws_elb.TargetDescription{
				Id:               aws_pkg.String(node.IPAddrs[0]),
				Port:             aws_pkg.Int64(443),
				AvailabilityZone: p.lb.AvailabilityZones[0].ZoneName,
			})
		}
	}

	elb := aws_elb.New(p.ses)
	input := &aws_elb.RegisterTargetsInput{
		TargetGroupArn: p.apiGrp.TargetGroupArn,
		Targets:        targets,
	}
	if _, err := elb.RegisterTargetsWithContext(ctx, input); err != nil {
		return errors.Wrap(err, "error registering API targets")
	}

	p.cfg.APIEndpoints = []config.ServiceEndpoint{
		{
			Addr: *p.lb.DNSName,
			Port: 443,
		},
	}

	return nil
}

func (p *awsProvider) registerTargetsSSH(ctx context.Context) error {
	elb := aws_elb.New(p.ses)

	input := &aws_elb.RegisterTargetsInput{
		TargetGroupArn: p.sshGrp.TargetGroupArn,
		Targets: []*aws_elb.TargetDescription{
			{
				Id:               aws_pkg.String(p.cfg.Nodes[0].IPAddrs[0]),
				Port:             aws_pkg.Int64(22),
				AvailabilityZone: p.lb.AvailabilityZones[0].ZoneName,
			},
		},
	}

	if _, err := elb.RegisterTargetsWithContext(ctx, input); err != nil {
		return errors.Wrap(err, "error registering SSH target")
	}

	p.cfg.Nodes[0].SSH.Addr = *p.lb.DNSName
	for i := range p.cfg.Nodes[1:] {
		p.cfg.Nodes[i].SSH.Addr = p.cfg.Nodes[i].IPAddrs[0]
		p.cfg.Nodes[i].SSH.ProxyHost = &p.cfg.Nodes[0].SSH
	}

	return nil
}

func (p *awsProvider) destroyLoadBalancer(ctx context.Context) error {
	elb := aws_elb.New(p.ses)

	input := &aws_elb.DescribeLoadBalancersInput{
		Names: []*string{aws_pkg.String(p.cfg.ID)},
	}
	output, _ := elb.DescribeLoadBalancersWithContext(ctx, input)
	if output != nil && len(output.LoadBalancers) > 0 {
		loadBalancerArn := output.LoadBalancers[0].LoadBalancerArn
		if _, err := elb.DeleteLoadBalancerWithContext(
			ctx, &aws_elb.DeleteLoadBalancerInput{
				LoadBalancerArn: loadBalancerArn,
			}); err != nil {
			return errors.Wrapf(
				err,
				"error deleting load balancer %s, arn=%s",
				p.cfg.ID,
				*loadBalancerArn)
		}
	}

	return nil
}

func (p *awsProvider) destoryTargetGroups(ctx context.Context) error {
	elb := aws_elb.New(p.ses)
	input := &aws_elb.DescribeTargetGroupsInput{
		Names: []*string{
			aws_pkg.String(fmt.Sprintf("%s-api", p.cfg.ID)),
			aws_pkg.String(fmt.Sprintf("%s-ssh", p.cfg.ID)),
		},
	}
	output, _ := elb.DescribeTargetGroupsWithContext(ctx, input)
	if output != nil {
		for i := range output.TargetGroups {
			targetGroupArn := output.TargetGroups[i].TargetGroupArn
			if _, err := elb.DeleteTargetGroupWithContext(
				ctx, &aws_elb.DeleteTargetGroupInput{
					TargetGroupArn: targetGroupArn,
				}); err != nil {
				return errors.Wrapf(
					err,
					"error deleting target group %s, arn=%s",
					p.cfg.ID,
					*targetGroupArn)
			}
		}
	}

	return nil
}

// Create the security group.
/*{
	{
		input := &aws_ec2.CreateSecurityGroupInput{
			GroupName:   aws_pkg.String(cfg.ID),
			VpcId:       aws_pkg.String(cfg.AWS.VpcID),
			Description: aws_pkg.String("Allows SSH & API server"),
		}
		output, err := p.ec2.CreateSecurityGroupWithContext(ctx, input)
		if err != nil {
			return nil, errors.Wrap(err, "error creating security group")
		}
		lb.secGrp = output.GroupId
	}
	{
		input := &aws_ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:   p.secGrp,
			GroupName: aws_pkg.String(cfg.ID),
			IpPermissions: []*aws_ec2.IpPermission{
				&aws_ec2.IpPermission{
					FromPort: aws_pkg.Int64(443),
					ToPort:   aws_pkg.Int64(443),
					IpRanges: []*aws_ec2.IpRange{
						&aws_ec2.IpRange{
							CidrIp:      aws_pkg.String("0.0.0.0/0"),
							Description: aws_pkg.String("Public HTTPS"),
						},
					},
					IpProtocol: aws_pkg.String(aws_ec2.TransportProtocolTcp),
				},
				&aws_ec2.IpPermission{
					FromPort: aws_pkg.Int64(2200),
					ToPort:   aws_pkg.Int64(2200 + int64(len(cfg.Nodes))),
					IpRanges: []*aws_ec2.IpRange{
						&aws_ec2.IpRange{
							CidrIp:      aws_pkg.String("0.0.0.0/0"),
							Description: aws_pkg.String("Public SSH"),
						},
					},
					IpProtocol: aws_pkg.String(aws_ec2.TransportProtocolTcp),
				},
			},
		}

		if _, err := p.ec2.AuthorizeSecurityGroupIngressWithContext(
			ctx, input); err != nil {
			return nil, errors.Wrap(
				err, "error creating security group ingress rules")
		}
	}
}*/
