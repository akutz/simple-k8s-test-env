// +build !ignore_autogenerated

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
// Code generated by deepcopy-gen. DO NOT EDIT.

package config

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
	pkgconfig "vmware.io/sk8/pkg/config"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSLoadBalancerConfig) DeepCopyInto(out *AWSLoadBalancerConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSLoadBalancerConfig.
func (in *AWSLoadBalancerConfig) DeepCopy() *AWSLoadBalancerConfig {
	if in == nil {
		return nil
	}
	out := new(AWSLoadBalancerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSLoadBalancerConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSLoadBalancerStatus) DeepCopyInto(out *AWSLoadBalancerStatus) {
	clone := in.DeepCopy()
	*out = *clone
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterProviderConfig) DeepCopyInto(out *ClusterProviderConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.NAT != nil {
		in, out := &in.NAT, &out.NAT
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
	out.OVA = in.OVA
	in.SSH.DeepCopyInto(&out.SSH)
	if in.CloudProvider != nil {
		in, out := &in.CloudProvider, &out.CloudProvider
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterProviderConfig.
func (in *ClusterProviderConfig) DeepCopy() *ClusterProviderConfig {
	if in == nil {
		return nil
	}
	out := new(ClusterProviderConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterProviderConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterStatus) DeepCopyInto(out *ClusterStatus) {
	clone := in.DeepCopy()
	*out = *clone
	return
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterStatus) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeployOVAConfig) DeepCopyInto(out *DeployOVAConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeployOVAConfig.
func (in *DeployOVAConfig) DeepCopy() *DeployOVAConfig {
	if in == nil {
		return nil
	}
	out := new(DeployOVAConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalCloudProviderConfig) DeepCopyInto(out *ExternalCloudProviderConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.Templates = in.Templates
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalCloudProviderConfig.
func (in *ExternalCloudProviderConfig) DeepCopy() *ExternalCloudProviderConfig {
	if in == nil {
		return nil
	}
	out := new(ExternalCloudProviderConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExternalCloudProviderConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalCloudProviderTemplates) DeepCopyInto(out *ExternalCloudProviderTemplates) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalCloudProviderTemplates.
func (in *ExternalCloudProviderTemplates) DeepCopy() *ExternalCloudProviderTemplates {
	if in == nil {
		return nil
	}
	out := new(ExternalCloudProviderTemplates)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImportOVAConfig) DeepCopyInto(out *ImportOVAConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImportOVAConfig.
func (in *ImportOVAConfig) DeepCopy() *ImportOVAConfig {
	if in == nil {
		return nil
	}
	out := new(ImportOVAConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InternalCloudProviderConfig) DeepCopyInto(out *InternalCloudProviderConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.Templates = in.Templates
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InternalCloudProviderConfig.
func (in *InternalCloudProviderConfig) DeepCopy() *InternalCloudProviderConfig {
	if in == nil {
		return nil
	}
	out := new(InternalCloudProviderConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *InternalCloudProviderConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InternalCloudProviderTemplates) DeepCopyInto(out *InternalCloudProviderTemplates) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InternalCloudProviderTemplates.
func (in *InternalCloudProviderTemplates) DeepCopy() *InternalCloudProviderTemplates {
	if in == nil {
		return nil
	}
	out := new(InternalCloudProviderTemplates)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineProviderConfig) DeepCopyInto(out *MachineProviderConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.Network.DeepCopyInto(&out.Network)
	out.OVA = in.OVA
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineProviderConfig.
func (in *MachineProviderConfig) DeepCopy() *MachineProviderConfig {
	if in == nil {
		return nil
	}
	out := new(MachineProviderConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MachineProviderConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineStatus) DeepCopyInto(out *MachineStatus) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.SSH != nil {
		in, out := &in.SSH, &out.SSH
		*out = new(pkgconfig.SSHEndpoint)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineStatus.
func (in *MachineStatus) DeepCopy() *MachineStatus {
	if in == nil {
		return nil
	}
	out := new(MachineStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MachineStatus) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
