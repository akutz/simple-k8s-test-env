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

	"vmware.io/sk8/pkg/config"
)

// MachineProviderConfig describes the information required to provision or
// reconcile a Machine.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MachineProviderConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`

	// An object missing this field at runtime is invalid.
	// +optional
	Datacenter string `json:"datacenter,omitempty"`

	// An object missing this field at runtime is invalid.
	// +optional
	Datastore string `json:"datastore,omitempty"`

	// An object missing this field at runtime is invalid.
	// +optional
	Folder string `json:"folder,omitempty"`

	// An object missing this field at runtime is invalid.
	// +optional
	ResourcePool string                      `json:"resourcePool,omitempty"`
	Network      config.MachineNetworkConfig `json:"network"`

	// OVA defines the content library item or template name used to deploy
	// machines.
	OVA DeployOVAConfig `json:"ova"`
}
