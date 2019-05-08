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

const (
	// GroupName is sk8's root group name.
	GroupName = "sk8.vmware.io"

	// KubernetesBuildIDLabelName is the label that indicates the build ID.
	KubernetesBuildIDLabelName = GroupName + "/kubernetes-build-id"

	// KubernetesBuildURLLabelName is the label that indicates the URL of
	// the build artifacts.
	KubernetesBuildURLLabelName = GroupName + "/kubernetes-build-url"

	// MachineRoleLabelName is the label that indicates the role of the
	// kubelet installed on a machine. Multiple roles may be specified
	// as comma-separated values.
	MachineRoleLabelName = GroupName + "/cluster-role"

	// SSHEndpointLabelName is the label that indicates the SSH endpoint
	// that may be used to access a machine in the cluster. That machine
	// should be able to access other machines in the cluster via SSH as
	// as bastion host.
	SSHEndpointLabelName = GroupName + "/ssh-endpoint"

	// ConfigDirLabelName is the directory to which providers may write
	// kubeconfig, sshconfig, SSH key files, etc.
	ConfigDirLabelName = GroupName + "/config-dir"

	// CloudProviderLabelName is the name of the label that points to
	// the cloud provider to configure for the cluster.
	CloudProviderLabelName = GroupName + "/cloud-provider"

	// PodNetworkCidrLabelName is the name of the label that indicates
	// the pod network cidr allocated to the cluster
	PodNetworkCidrLabelName = GroupName + "/pod-network-cidr"

	// Version is the sk8-wide version.
	Version = "v1alpha0"
)
