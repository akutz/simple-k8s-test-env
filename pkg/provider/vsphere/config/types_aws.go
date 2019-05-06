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
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWSLoadBalancerConfig describes how to access machines via an AWS
// load balancer.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AWSLoadBalancerConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`

	// +optional
	AccessKeyID string `json:"accessKeyID,omitempty"`
	// +optional
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	// +optional
	Region string `json:"region,omitempty"`
	// +optional
	MaxRetries int32 `json:"maxRetries,omitempty"`

	// Defaults to 8888
	// +optional
	HealthCheckPort int32 `json:"healthCheckPort,omitempty"`

	// SubnetID is the ID of the subnet connected to VMC.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	SubnetID string `json:"subnetID,omitempty"`

	// VpcID is the ID of the VPC connected to VMC.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	VpcID string `json:"vpcID,omitempty"`
}

// SetDefaults_AWSLoadBalancerConfig sets uninitialized fields to their default value.
func SetDefaults_AWSLoadBalancerConfig(obj *AWSLoadBalancerConfig) {
	if obj.AccessKeyID == "" {
		obj.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if obj.SecretAccessKey == "" {
		obj.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	if obj.Region == "" {
		obj.Region = os.Getenv("AWS_DEFAULT_REGION")
		if obj.Region == "" {
			obj.Region = os.Getenv("AWS_REGION")
		}
	}
	if obj.HealthCheckPort == 0 {
		obj.HealthCheckPort = 8888
	}
}
