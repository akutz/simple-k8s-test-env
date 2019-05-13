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

// ExternalCloudProviderConfig is the cloud provider configuration for the
// external vSphere cloud provider.
//
// The Templates field may be used to specify custom templates for the
// external cloud provider. If a required template is omitted then the
// default template is used.
//
// Whether custom or default, all templates are processed with the following
// functions in the function map:
//   * base64  returns a base64 encoded string
//   * join    joins a list of strings with a comma
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ExternalCloudProviderConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`

	// ConfigMapName is the name of the config map that contains the cloud
	// provider configuration.
	//
	// Defaults to "cloud-config-volume"
	// +optional
	ConfigMapName string `json:"configMapName,omitempty"`

	// ConfigMapName is the name of the config map that contains the cloud
	// provider configuration.
	//
	// Defaults to "cloud.conf"
	// +optional
	ConfigFileName string `json:"configFileName,omitempty"`

	// Datacenter is the datacenter where Kubernetes machines are created.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Datacenter string `json:"datacenter,omitempty"`

	// Image is the cloud provider image to deploy.
	//
	// Defaults to "gcr.io/cloud-provider-vsphere/vsphere-cloud-controller-manager:latest"
	// +optional
	Image string `json:"image,omitempty"`

	// Insecure is a flag that controls whether or not to validate the
	// remote endpoint's certificate.
	//
	// Defaults to the analogue value in the ClusterProviderConfig.
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// Namespace is the Kubernetes namespace into which the cloud provider
	// and its related resources are deployed.
	//
	// Defaults to "kube-system"
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Password is used to connect to the vSphere server.
	//
	// Defaults to the analogue value in the ClusterProviderConfig.
	// +optional
	Password string `json:"password,omitempty"`

	// Region is used to configure zone support.
	// +optional
	Region string `json:"region,omitempty"`

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

	// ServiceAccount is the name of the Kubernetes service account used
	// by the cloud provider.
	//
	// Defaults to "cloud-controller-manager"
	// +optional
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// ServicePort is the port on which the cloud provider service listens.
	//
	// Defaults to 43001
	// +optional
	ServicePort int32 `json:"servicePort,omitempty"`

	// Username is used to connect to the vSphere server.
	//
	// Defaults to the analogue value in the ClusterProviderConfig.
	// +optional
	Username string `json:"username,omitempty"`

	// Zone is used to configure zone support.
	// +optional
	Zone string `json:"zone,omitempty"`

	// Templates are the templates used to configure the cloud provider.
	Templates ExternalCloudProviderTemplates `json:"templates"`
}

// SetDefaults_ExternalCloudProviderConfig sets uninitialized fields to their default value.
func SetDefaults_ExternalCloudProviderConfig(obj *ExternalCloudProviderConfig) {
	if obj.ConfigMapName == "" {
		obj.ConfigMapName = "cloud-config-volume"
	}
	if obj.ConfigFileName == "" {
		obj.ConfigFileName = "cloud.conf"
	}
	if obj.Image == "" {
		obj.Image = extCCMDefaultImage
	}
	if obj.Namespace == "" {
		obj.Namespace = "kube-system"
	}
	if obj.SecretName == "" {
		obj.SecretName = "vccm"
	}
	if obj.ServiceAccount == "" {
		obj.ServiceAccount = "cloud-controller-manager"
	}
	if obj.ServicePort == 0 {
		obj.ServicePort = 43001
	}
}

// ExternalCloudProviderTemplates are the templates used to configure the
// cloud provider.
type ExternalCloudProviderTemplates struct {
	Config         string `json:"config,omitempty"`
	Deployment     string `json:"deployment,omitempty"`
	Roles          string `json:"roles,omitempty"`
	RoleBindings   string `json:"roleBindings,omitempty"`
	Secrets        string `json:"secrets,omitempty"`
	Service        string `json:"service,omitempty"`
	ServiceAccount string `json:"serviceAccount,omitempty"`
}

// SetDefaults_ExternalCloudProviderTemplates sets uninitialized fields to their default value.
func SetDefaults_ExternalCloudProviderTemplates(obj *ExternalCloudProviderTemplates) {
	if len(obj.Config) == 0 {
		obj.Config = extCCMConfigFormat
	}
	if len(obj.Deployment) == 0 {
		obj.Deployment = extCCMDeployPodFormat
	}
	if len(obj.Roles) == 0 {
		obj.Roles = extCCMRolesFormat
	}
	if len(obj.RoleBindings) == 0 {
		obj.RoleBindings = extCCMRoleBindingsFormat
	}
	if len(obj.Secrets) == 0 {
		obj.Secrets = extCCMSecretsFormat
	}
	if len(obj.Service) == 0 {
		obj.Service = extCCMServiceFormat
	}
	if len(obj.ServiceAccount) == 0 {
		obj.ServiceAccount = extCCMServiceAccountFormat
	}
}
