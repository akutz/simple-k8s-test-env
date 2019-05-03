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

// DeployOVAConfig describes the template or content library item used
// to deploy VMs.
type DeployOVAConfig struct {
	// Method is the method used to deploy the cluster node VMs.
	//
	// Defaults to "content-library"
	//
	// +optional
	Method string `json:"method,omitempty"`

	// Source is the path to the content library item or template used to
	// deploy the cluster node VMs.
	//
	// Defaults to "/sk8/photon3-cloud-init"
	//
	// +optional
	Source string `json:"source,omitempty"`
}

// SetDefaults_DeployOVAConfig sets uninitialized fields to their default value.
func SetDefaults_DeployOVAConfig(obj *DeployOVAConfig) {
	if obj.Method == "" {
		obj.Method = "content-library"
	}
	if obj.Source == "" {
		obj.Source = "/sk8/photon3-cloud-init"
	}
}

// ImportOVAConfig describes the OVA used to deploy VMs and how to import
// it into a content library or as a template.
type ImportOVAConfig struct {
	// Method is the type of import used.
	//
	// Defaults to "content-library"
	//
	// +optional
	Method string `json:"method,omitempty"`

	// Source is the URL of the OVA to be imported as a template or content
	// library item and used to deploy VMs for cluster nodes.
	//
	// The OVA must have cloud-init and the VMware GuestInfo datasource enabled.
	//
	// Defaults to https://s3-us-west-2.amazonaws.com/cnx.vmware/photon3-cloud-init.ova
	//
	// +optional
	Source string `json:"source,omitempty"`

	// Target is the path to a content library item or template that is
	// the target of the OVA import.
	//
	// Defaults to "/sk8/photon3-cloud-init"
	//
	// +optional
	Target string `json:"target,omitempty"`
}

// SetDefaults_ImportOVAConfig sets uninitialized fields to their default value.
func SetDefaults_ImportOVAConfig(obj *ImportOVAConfig) {
	if obj.Method == "" {
		obj.Method = "content-library"
	}
	if obj.Source == "" {
		obj.Source = "https://s3-us-west-2.amazonaws.com/cnx.vmware/photon3-cloud-init.ova"
	}
	if obj.Target == "" {
		obj.Target = "/sk8/photon3-cloud-init"
	}
}
