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

package aws

import (
	"context"
	"fmt"
	"time"

	aws_pkg "github.com/aws/aws-sdk-go/aws"
	aws_credentials "github.com/aws/aws-sdk-go/aws/credentials"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_elb "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
)

// NewSession returns a new AWS session.
func NewSession(
	ctx context.Context,
	nat *vconfig.AWSLoadBalancerConfig) (*aws_session.Session, error) {

	ses, err := aws_session.NewSession(&aws_pkg.Config{
		Region:     aws_pkg.String(nat.Region),
		MaxRetries: aws_pkg.Int(int(nat.MaxRetries)),
		Credentials: aws_credentials.NewChainCredentials(
			[]aws_credentials.Provider{
				&aws_credentials.StaticProvider{
					Value: aws_credentials.Value{
						AccessKeyID:     nat.AccessKeyID,
						SecretAccessKey: nat.SecretAccessKey,
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
	return ses, nil
}

// FindLoadBalancer finds the load balancer that matches the provided
// clusterName.
func FindLoadBalancer(
	ctx context.Context,
	ses *aws_session.Session,
	nat *vconfig.AWSLoadBalancerConfig,
	clusterName string) (*aws_elb.LoadBalancer, error) {

	elb := aws_elb.New(ses)
	input := &aws_elb.DescribeLoadBalancersInput{
		Names: []*string{aws_pkg.String(clusterName)},
	}
	output, err := elb.DescribeLoadBalancersWithContext(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "error finding existing load balancer")
	}
	return output.LoadBalancers[0], nil
}

// FindTargetGroups returns the API and SSH target groups for the given
// loadBalancer.
func FindTargetGroups(
	ctx context.Context,
	ses *aws_session.Session,
	nat *vconfig.AWSLoadBalancerConfig,
	clusterName string,
	lb *aws_elb.LoadBalancer) (
	tgAPI *aws_elb.TargetGroup, tgSSH *aws_elb.TargetGroup, err error) {

	elb := aws_elb.New(ses)
	input := &aws_elb.DescribeTargetGroupsInput{
		LoadBalancerArn: lb.LoadBalancerArn,
	}
	output, err2 := elb.DescribeTargetGroupsWithContext(ctx, input)
	if err2 != nil {
		return nil, nil, errors.Wrapf(
			err2,
			"error describing target groups for lb %q",
			*lb.LoadBalancerName)
	}

	for _, tg := range output.TargetGroups {
		switch *tg.TargetGroupName {
		case fmt.Sprintf("%s-api", clusterName):
			tgAPI = tg
		case fmt.Sprintf("%s-ssh", clusterName):
			tgSSH = tg
		}
	}

	return tgAPI, tgSSH, nil
}

// WaitLoadBalancer blocks until the the load balancer provisioning operation
// has completed or the context is cancelled. Please note that a completed
// provision operation does not mean the operation completed successfully.
func WaitLoadBalancer(
	ctx context.Context,
	ses *aws_session.Session,
	lb *aws_elb.LoadBalancer) error {

	elb := aws_elb.New(ses)
	done := make(chan struct{})
	errs := make(chan error)
	input := &aws_elb.DescribeLoadBalancersInput{
		LoadBalancerArns: []*string{lb.LoadBalancerArn},
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

			lb.State = output.LoadBalancers[0].State
			lb.DNSName = output.LoadBalancers[0].DNSName
			code := *lb.State.Code
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
