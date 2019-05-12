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
	"os"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"vmware.io/sk8/pkg/config"
)

// ClusterProviderConfig describes the endpoint and credentials used to access
// a vSphere system.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterProviderConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`

	// Server is the address of the vSphere endpoint.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Server string `json:"server,omitempty"`

	// Username is the name used to log into the vSphere server.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Username string `json:"username,omitempty"`

	// Password is the password used to log into the vSphere server.
	//
	// An object missing this field at runtime is invalid.
	// +optional
	Password string `json:"password,omitempty"`

	// Insecure is a flag that controls whether or not to validate the
	// vSphere server's certificate.
	Insecure bool `json:"insecure,omitempty"`

	// NAT is the configuration for the service that enables external access to
	// machines deployed to a private network. The currently supported types
	// are:
	//   * sk8.vmware.io/LinuxVirtualSwitchConfig
	//   * vsphere.sk8.vmware.io/AWSLoadBalancerConfig
	// +optional
	NAT *runtime.RawExtension `json:"nat,omitempty"`

	// Net contains the network manifest(s) applied to the cluster.
	//
	// Defaults to a WeaveWorks network.
	Net string `json:"net,omitempty"`

	// OVA describes the OVA used to deploy machines and whether to import
	// it as a content library item or template.
	OVA ImportOVAConfig `json:"ova"`

	// SSH defines the SSH user and private key used to access the cluster's
	// machines.
	SSH config.SSHCredential `json:"ssh"`

	// CloudProvider is the configuration for the cloud provider. The
	// currently supported types are:
	//   * vsphere.sk8.vmware.io/ExternalCloudProviderConfig
	//   * vsphere.sk8.vmware.io/InternalCloudProviderConfig
	CloudProvider *runtime.RawExtension `json:"cloudProvider,omitempty"`
}

// UnmarshalJSON ensures that the object's NAT field is unmarshaled at
// the same time as the parent object.
func (c *ClusterProviderConfig) UnmarshalJSON(data []byte) error {

	// Create an anonymous struct into which the provided data is
	// unmarshaled. The struct is initialized with the data from the
	// parent ClusterProviderConfig.
	var obj = struct {
		Server        string                `json:"server,omitempty"`
		Username      string                `json:"username,omitempty"`
		Password      string                `json:"password,omitempty"`
		CloudProvider *runtime.RawExtension `json:"cloudProvider,omitempty"`
		NAT           *runtime.RawExtension `json:"nat,omitempty"`
		Net           string                `json:"net,omitempty"`
		OVA           ImportOVAConfig       `json:"ova,omitempty"`
		SSH           config.SSHCredential  `json:"ssh,omitempty"`
	}{
		Server:   c.Server,
		Username: c.Username,
		Password: c.Password,
		Net:      c.Net,
	}
	c.OVA.DeepCopyInto(&obj.OVA)
	c.SSH.DeepCopyInto(&obj.SSH)

	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	// If there is new CloudProvider data then it needs to be unmarshaled into
	// either the config's existing CloudProvider object or a new object.
	if obj.CloudProvider != nil && len(obj.CloudProvider.Raw) > 0 {
		log.Debug("is ccm data")
		var cfg runtime.Object

		// Does the ClusterProviderConfig already have a CloudProvider field that
		// contains a valid Object? If so, re-use it as the target of
		// the unmarshal.
		//
		// Otherwise the data is unmarshaled in order to determine what
		// kind of CloudProvider implementation to create.
		if c.CloudProvider != nil && c.CloudProvider.Object != nil {
			log.Debug("re-use ccm object")
			// Re-use the existing CloudProvider object.
			cfg = c.CloudProvider.Object
		} else {
			log.Debug("create ccm object")
			// Inspect the data to determine the type CloudProvider object to
			// create.
			var typeMeta runtime.TypeMeta
			if err := yaml.Unmarshal(obj.CloudProvider.Raw, &typeMeta); err != nil {
				return err
			}
			log.WithField("typeMeta", typeMeta).Debug("discovered api object")
			switch typeMeta.GroupVersionKind().Kind {
			case "ExternalCloudProviderConfig":
				cfg = &ExternalCloudProviderConfig{}
			}
		}

		// If the CloudProvider config is not nil, then the given CloudProvider
		// data should be unmarshaled into the object.
		if cfg != nil {
			log.Debug("unmarshal raw api object into external cloud provider")
			if err := yaml.Unmarshal(obj.CloudProvider.Raw, cfg); err != nil {
				return err
			}
			obj.CloudProvider.Raw = nil
			obj.CloudProvider.Object = cfg
		}
	}

	// If there is new NAT data then it needs to be unmarshaled into either
	// the config's existing NAT object or a new object.
	if obj.NAT != nil && len(obj.NAT.Raw) > 0 {
		var cfg runtime.Object

		// Does the ClusterProviderConfig already have a NAT field that
		// contains a valid Object? If so, re-use it as the target of
		// the unmarshal.
		//
		// Otherwise the data is unmarshaled in order to determine what
		// kind of NAT implementation to create.
		if c.NAT != nil && c.NAT.Object != nil {
			// Re-use the existing NAT object.
			cfg = c.NAT.Object
		} else {
			// Inspect the data to determine the type of NAT object to
			// create.
			var typeMeta runtime.TypeMeta
			if err := yaml.Unmarshal(obj.NAT.Raw, &typeMeta); err != nil {
				return err
			}
			switch typeMeta.GroupVersionKind().Kind {
			case "LinuxVirtualSwitchConfig":
				cfg = &config.LinuxVirtualSwitchConfig{}
			case "AWSLoadBalancerConfig":
				cfg = &AWSLoadBalancerConfig{}
			}
		}

		// If the NAT config is not nil, then the given NAT data should be
		// unmarshaled into the object.
		if cfg != nil {
			if err := yaml.Unmarshal(obj.NAT.Raw, cfg); err != nil {
				return err
			}
			obj.NAT.Raw = nil
			obj.NAT.Object = cfg
		}
	}

	c.SetGroupVersionKind(SchemeGroupVersion.WithKind("ClusterProviderConfig"))
	c.Server = obj.Server
	c.Username = obj.Username
	c.Password = obj.Password
	if obj.CloudProvider != nil {
		c.CloudProvider = obj.CloudProvider
	}
	if obj.NAT != nil {
		c.NAT = obj.NAT
	}
	c.Net = obj.Net
	obj.OVA.DeepCopyInto(&c.OVA)
	obj.SSH.DeepCopyInto(&c.SSH)

	return nil
}

// SetDefaults_ClusterProviderConfig sets uninitialized fields to their default value.
func SetDefaults_ClusterProviderConfig(obj *ClusterProviderConfig) {
	if obj.Server == "" {
		obj.Server = os.Getenv("SK8_VSPHERE_SERVER")
		if obj.Server == "" {
			obj.Server = os.Getenv("VSPHERE_SERVER")
		}
	}
	if obj.Username == "" {
		obj.Username = os.Getenv("SK8_VSPHERE_USERNAME")
		if obj.Username == "" {
			obj.Username = os.Getenv("VSPHERE_USERNAME")
			if obj.Username == "" {
				obj.Username = os.Getenv("VSPHERE_USER")
			}
		}
	}
	if obj.Password == "" {
		obj.Password = os.Getenv("SK8_VSPHERE_PASSWORD")
		if obj.Password == "" {
			obj.Password = os.Getenv("VSPHERE_PASSWORD")
		}
	}

	config.SetDefaults_SSHCredential(&obj.SSH)

	if obj.CloudProvider != nil {
		switch cfg := obj.CloudProvider.Object.(type) {
		case *ExternalCloudProviderConfig:
			if cfg.ServerAddr == "" {
				cfg.ServerAddr = obj.Server
			}
			if cfg.ServerPort == 0 {
				cfg.ServerPort = 443
			}
			if cfg.Username == "" {
				cfg.Username = obj.Username
			}
			if cfg.Password == "" {
				cfg.Password = obj.Password
			}
			SetObjectDefaults_ExternalCloudProviderConfig(cfg)
		}
	}

	if obj.NAT != nil {
		switch cfg := obj.NAT.Object.(type) {
		case *config.LinuxVirtualSwitchConfig:
			config.SetObjectDefaults_LinuxVirtualSwitchConfig(cfg)
		case *AWSLoadBalancerConfig:
			SetObjectDefaults_AWSLoadBalancerConfig(cfg)
		}
	}

	if obj.Net == "" {
		obj.Net = weaveWorksFormat
	}
}
