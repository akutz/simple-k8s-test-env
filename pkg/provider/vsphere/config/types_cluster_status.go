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

package config

import (
	"encoding/json"
	"sync"

	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_elb "github.com/aws/aws-sdk-go/service/elbv2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"vmware.io/sk8/pkg/config"
)

// ClusterStatus describes the status of a cluster.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterStatus struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`

	// KubeJoinCmd is the token used to join the cluster.
	KubeJoinCmd string `json:"kubeJoinCmd,omitempty"`

	// OVAID is the ID of the OVA to deploy.
	OVAID string `json:"ovaID,omitempty"`

	// SSH is the bastion host used to access the machines via SSH.
	SSH *config.SSHEndpoint `json:"ssh,omitempty"`

	// AWS is the status of the AWSLoadBalancer NAT provider
	AWS *AWSLoadBalancerStatus `json:"aws,omitempty"`

	// ControlPlaneEndpoint is the address of the host on which the control
	// plane was initialized.
	ControlPlaneEndpoint string `json:"controlPlaneEndpoint,omitempty"`

	// ControlPlaneCanOwn is a channel used to signal that a machine is
	// responsible for initializing the control plane.
	ControlPlaneCanOwn chan struct{} `json:"-"`

	// ControlPlaneOnline is a channel that is closed once the control
	// plane is online.
	ControlPlaneOnline chan struct{} `json:"-"`

	// SSHConfigMu controls access to the local SSH config file.
	SSHConfigMu sync.Mutex `json:"-"`
}

func (in *ClusterStatus) DeepCopy() *ClusterStatus {
	if in == nil {
		return nil
	}
	out := &ClusterStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: in.TypeMeta.APIVersion,
			Kind:       in.TypeMeta.Kind,
		},
		ControlPlaneCanOwn: make(chan struct{}, 1),
		ControlPlaneOnline: make(chan struct{}),
		KubeJoinCmd:        in.KubeJoinCmd,
		OVAID:              in.OVAID,
	}
	out.ControlPlaneCanOwn <- struct{}{}
	if in.SSH != nil {
		out.SSH = &config.SSHEndpoint{}
		in.SSH.DeepCopyInto(out.SSH)
	}
	if in.AWS != nil {
		out.AWS = &AWSLoadBalancerStatus{}
		in.AWS.DeepCopyInto(out.AWS)
	}
	return out
}

// AWSLoadBalancerStatus contains information about the AWSLoadBalancer
// NAT provider.
type AWSLoadBalancerStatus struct {
	Session      *aws_session.Session  `json:"-"`
	LoadBalancer *aws_elb.LoadBalancer `json:"-"`
	API          *aws_elb.TargetGroup  `json:"-"`
	SSH          *aws_elb.TargetGroup  `json:"-"`
	CanOwn       chan struct{}         `json:"-"`
	Online       chan struct{}         `json:"-"`
}

func (in *AWSLoadBalancerStatus) DeepCopy() *AWSLoadBalancerStatus {
	out := &AWSLoadBalancerStatus{
		Session:      in.Session,
		LoadBalancer: in.LoadBalancer,
		API:          in.API,
		SSH:          in.SSH,
		CanOwn:       make(chan struct{}, 1),
		Online:       make(chan struct{}),
	}
	out.CanOwn <- struct{}{}
	return out
}

func (in *AWSLoadBalancerStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		LoadBalancerARN   *string `json:"loadBalancerARN,omitempty"`
		LoadBalancerDNS   *string `json:"loadBalancerDNS,omitempty"`
		APITargetGroupARN *string `json:"apiTargetGroupARN,omitempty"`
		SSHTargetGroupARN *string `json:"sshTargetGroupARN,omitempty"`
	}{
		LoadBalancerARN:   in.LoadBalancer.LoadBalancerArn,
		LoadBalancerDNS:   in.LoadBalancer.DNSName,
		APITargetGroupARN: in.API.TargetGroupArn,
		SSHTargetGroupARN: in.SSH.TargetGroupArn,
	})
}

/*type SSHStatus struct {
	sync.Mutex

	// Bastion is the bastion host used to access the machines via SSH.
	Bastion *config.SSHEndpoint `json:"ssh"`

	// SSHConfigMu controls access to the local SSH config file.
	SSHConfigMu sync.Mutex `json:"-"`
}*/
