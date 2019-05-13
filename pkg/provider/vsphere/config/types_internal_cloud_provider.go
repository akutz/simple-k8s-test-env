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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InternalCloudProviderConfig is the cloud provider configuration for the
// internal vSphere cloud provider.
//
// The Templates field may be used to specify custom templates for the
// internal cloud provider. If a required template is omitted then the
// default template is used.
//
// Whether custom or default, all templates are processed with the following
// functions in the function map:
//   * base64  returns a base64 encoded string
//   * join    joins a list of strings with a comma
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type InternalCloudProviderConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`

	// ConfigFilePath is the path on worker nodes where the cloud provider's
	// configuration file is written.
	//
	// Defaults to "/etc/kubernetes/cloud.conf"
	// +optional
	ConfigFilePath string `json:"configFilePath,omitempty"`

	// Datacenter is the datacenter where Kubernetes machines are deployed.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Datacenter string `json:"datacenter,omitempty"`

	// Datastore is the datastore where Kubernetes machines are deployed.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Datastore string `json:"datastore,omitempty"`

	// Folder is the folder where Kubernetes machines are deployed.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Folder string `json:"folder,omitempty"`

	// Insecure is a flag that controls whether or not to validate the
	// remote endpoint's certificate.
	//
	// Defaults to the analogue value in the ClusterProviderConfig.
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// Namespace is the Kubernetes namespace into which the cloud provider's
	// secret is deployed.
	//
	// Defaults to "kube-system"
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Network is the network used by the Kubernetes machines.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Network string `json:"network,omitempty"`

	// Password is used to connect to the vSphere server.
	//
	// Defaults to the analogue value in the ClusterProviderConfig.
	// +optional
	Password string `json:"password,omitempty"`

	// ResourcePool is the resource pool where Kubernetes machines are deployed.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	ResourcePool string `json:"resourcePool,omitempty"`

	// Region is used to configure zone support.
	// +optional
	Region string `json:"region,omitempty"`

	// Region is used to configure zone support.
	//
	// Defaults to "pvscsi"
	// +optional
	SCSIControllerType string `json:"scsiControllerType,omitempty"`

	// SecretName is the name of the Kubernetes secret that contains the
	// credentials used by the cloud provider.
	//
	// Defaults to "vccm"
	// +optional
	SecretName string `json:"secretName,omitempty"`

	// ServerAddr is the address of the vSphere server.
	//
	// Defaults to the analogue value in the ClusterProviderConfig.
	// +optional
	ServerAddr string `json:"serverAddr,omitempty"`

	// ServerPort is the port on which the vSphere server is listening.
	//
	// Defaults to the analogue value in the ClusterProviderConfig.
	// +optional
	ServerPort int32 `json:"serverPort,omitempty"`

	// Username is used to connect to the vSphere server.
	//
	// Defaults to the analogue value in the ClusterProviderConfig.
	// +optional
	Username string `json:"username,omitempty"`

	// Zone is used to configure zone support.
	// +optional
	Zone string `json:"zone,omitempty"`

	// Templates are the templates used to configure the cloud provider.
	Templates InternalCloudProviderTemplates `json:"templates"`
}

// SetDefaults_InternalCloudProviderConfig sets uninitialized fields to their default value.
func SetDefaults_InternalCloudProviderConfig(obj *InternalCloudProviderConfig) {
	if obj.ConfigFilePath == "" {
		obj.ConfigFilePath = "/etc/kubernetes/cloud.conf"
	}
	if obj.Namespace == "" {
		obj.Namespace = "kube-system"
	}
	if obj.SecretName == "" {
		obj.SecretName = "vccm"
	}
	if obj.SCSIControllerType == "" {
		obj.SCSIControllerType = "pvscsi"
	}
}

// InternalCloudProviderTemplates are the templates used to configure the
// cloud provider.
type InternalCloudProviderTemplates struct {
	Config  string `json:"config,omitempty"`
	Secrets string `json:"secrets,omitempty"`
}

// SetDefaults_InternalCloudProviderTemplates sets uninitialized fields to their default value.
func SetDefaults_InternalCloudProviderTemplates(obj *InternalCloudProviderTemplates) {
	if len(obj.Config) == 0 {
		obj.Config = intCCMConfigFormat
	}
	if len(obj.Secrets) == 0 {
		obj.Secrets = intCCMSecretsFormat
	}
}
