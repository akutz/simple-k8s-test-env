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

package cluster

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/yaml"

	"vmware.io/sk8/pkg/builds"
	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/config/encoding"
)

// Cluster describes the information required to turn up, reconcile, or
// delete a Kubernetes cluster.
type Cluster struct {
	// Cluster is the CAPI cluster.
	Cluster capi.Cluster

	// Machines is a list of the machines that belong to the cluster.
	Machines capi.MachineList
}

// New returns a new Cluster object.
func New(
	clusterConfig runtime.Object,
	machineConfigs ...runtime.Object) *Cluster {

	capiGV := capi.SchemeGroupVersion

	var c Cluster
	c.Cluster.SetGroupVersionKind(capiGV.WithKind("Cluster"))
	c.Machines.SetGroupVersionKind(capiGV.WithKind("MachineList"))
	encoding.Scheme.Default(&c.Cluster)
	encoding.Scheme.Default(&c.Machines)

	c.Cluster.SetLabels(map[string]string{})
	c.Cluster.Spec.ProviderSpec.Value = &runtime.RawExtension{
		Object: clusterConfig,
	}

	c.Machines.Items = make([]capi.Machine, len(machineConfigs))
	for i, machineConfig := range machineConfigs {
		machine := &c.Machines.Items[i]
		machine.SetLabels(map[string]string{})
		machine.SetGroupVersionKind(capiGV.WithKind("Machine"))
		machine.Spec.ProviderSpec.Value = &runtime.RawExtension{
			Object: machineConfig,
		}
	}

	return &c
}

// NewFromRoles returns a new Cluster object based on the provided
// machineRoles. Each element of machineRoles may be comma-separated in
// order to specify multiple roles for that machine.
func NewFromRoles(
	clusterProviderConfigGVK schema.GroupVersionKind,
	machineProviderConfigGVK schema.GroupVersionKind,
	machineRoles ...config.MachineRole) (*Cluster, error) {

	clusterProCfg, err := encoding.New(clusterProviderConfigGVK)
	if err != nil {
		return nil, err
	}

	machineProCfgs := make([]runtime.Object, len(machineRoles))
	for i := range machineRoles {
		obj, err := encoding.New(machineProviderConfigGVK)
		if err != nil {
			return nil, err
		}
		machineProCfgs[i] = obj
	}

	clu := New(clusterProCfg, machineProCfgs...)
	for i := range clu.Machines.Items {
		machine := &clu.Machines.Items[i]
		machine.Labels[config.MachineRoleLabelName] = machineRoles[i].String()
	}

	return clu, nil
}

// ApplyDefaultProviderConfigs updates a Cluster's ClusterProviderConfig
// and all of the MachineProviderConfigs with API objects from the
// provided data that match the GroupVersionKind of the existing provider
// config objects.
func (c *Cluster) ApplyDefaultProviderConfigs(rdrs ...io.Reader) error {
	var objs []runtime.Object

	for _, r := range rdrs {
		// Read all of the API objects from the provided data.
		yamlReader := utilyaml.NewYAMLReader(bufio.NewReader(r))
		for {
			buf, err := yamlReader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			// If the data can be decoded successfully then add it to the
			// list of objects that may include default config data for the
			// providers.
			if obj, err := encoding.Decode(buf); err == nil {
				objs = append(objs, obj)
			}
		}
	}

	// Apply the defaults to the ClusterProviderConfig.
	if err := c.applyDefaultProviderConfigs(
		c.Cluster.Spec.ProviderSpec.Value, objs...); err != nil {
		return err
	}

	// Apply the defaults to all of the MachineProviderConfigs.
	for i := range c.Machines.Items {
		m := &c.Machines.Items[i]
		if err := c.applyDefaultProviderConfigs(
			m.Spec.ProviderSpec.Value, objs...); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) applyDefaultProviderConfigs(
	val *runtime.RawExtension, objs ...runtime.Object) error {

	var (
		// allDefCfg is a combined view of all the default configuration data
		allDefCfg runtime.Object
		curCfgGVK = val.Object.GetObjectKind().GroupVersionKind()
	)

	// First combine all of the apllicable objects into cfgObj.
	for _, defCfg := range objs {
		if defCfg.GetObjectKind().GroupVersionKind() == curCfgGVK {
			if allDefCfg == nil {
				allDefCfg = defCfg
			} else {
				defCfgBuf, err := yaml.Marshal(defCfg)
				if err != nil {
					return err
				}
				if err := encoding.DecodeInto(defCfgBuf, allDefCfg); err != nil {
					return err
				}
			}
		}
	}

	// Now rebase the existing config data onto the combined, default config.
	if allDefCfg != nil {
		// newCfg is just an alias for allDefCfg in order to make the following
		// logic more self-descriptive
		newCfg := allDefCfg

		// Marshal the current provider config.
		curCfgBuf, err := yaml.Marshal(val.Object)
		if err != nil {
			return err
		}

		// Take the marshaled copy of the current provider config and decode it
		// into the newCfg object. This is used instead of DeepCopyInto because
		// the latter method overwrites non-empty fields with empty ones.
		if err := encoding.DecodeInto(curCfgBuf, newCfg); err != nil {
			return err
		}

		// Replace the current config with the new one and ensure the scheme
		// defaults are set.
		val.Object = newCfg
		encoding.Scheme.Default(val.Object)
	}

	return nil
}

// DeepCopy copies the Cluster.
func (c *Cluster) DeepCopy() *Cluster {
	var cc Cluster
	c.Cluster.DeepCopyInto(&cc.Cluster)
	c.Machines.DeepCopyInto(&cc.Machines)
	return &cc
}

// MarshalText returns the YAML representation of the object.
func (c *Cluster) MarshalText() ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, obj := range []runtime.Object{&c.Cluster, &c.Machines} {
		buf2, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(buf2); err != nil {
			return nil, err
		}
		buf.WriteString("\n\n---\n\n")
	}
	return buf.Bytes(), nil
}

// UnmarshalText unmarshals the provided YAML bytes into the object.
func (c *Cluster) UnmarshalText(data []byte) error {
	r := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
	for {
		buf, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		var typeMeta runtime.TypeMeta
		if err := yaml.Unmarshal(buf, &typeMeta); err != nil {
			return err
		}

		var obj runtime.Object
		switch typeMeta.GroupVersionKind() {
		case capi.SchemeGroupVersion.WithKind("Cluster"):
			obj = &c.Cluster
		case capi.SchemeGroupVersion.WithKind("MachineList"):
			obj = &c.Machines
		}

		if obj != nil {
			if err := yaml.Unmarshal(buf, obj); err != nil {
				return err
			}
			encoding.Scheme.Default(obj)
		}
	}

	// Decode the provider config from the raw data into a runtime.Object.
	if _, err := encoding.FromRaw(
		c.Cluster.Spec.ProviderSpec.Value); err != nil {
		return errors.Wrap(err, "Cluster")
	}

	for i, machine := range c.Machines.Items {
		// Decode the provider config from the raw data into a runtime.Object.
		if _, err := encoding.FromRaw(
			machine.Spec.ProviderSpec.Value); err != nil {
			return errors.Wrapf(err, "Machine[%d]", i)
		}
	}

	return nil
}

// WithCloudProvider configures the cluster to use the specified
// cloud provider.
func (c *Cluster) WithCloudProvider(cloudProvider string) *Cluster {
	if cloudProvider != "" {
		c.Cluster.Labels[config.CloudProviderLabelName] = cloudProvider
	}
	return c
}

// WithPodNetworkCidr configures a cluster with a pod network cidr
func (c *Cluster) WithPodNetworkCidr(podCidr string) *Cluster {
	if podCidr != "" {
		c.Cluster.Labels[config.PodNetworkCidrLabelName] = podCidr
	}

	return c
}

// WithDefaults updates the Cluster's ClusterProviderConfig object and all of
// the MachineProviderConfig objects using the default configuration data from
// the provided file paths.
func (c *Cluster) WithDefaults(paths ...string) (*Cluster, error) {
	rdrs := make([]io.Reader, len(paths))
	for i, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		rdrs[i] = f
		log.WithField("path", p).Debug("loading defaults")
	}
	if err := c.ApplyDefaultProviderConfigs(rdrs...); err != nil {
		return nil, err
	}
	return c, nil
}

// WithKubernetesVersion assigns the version of Kubernetes to deploy.
func (c *Cluster) WithKubernetesVersion(version string) *Cluster {
	return c.WithKubernetesBuildInfo("", "", version)
}

// WithKubernetesBuildID assigns the build of Kubernetes to deploy.
func (c *Cluster) WithKubernetesBuildID(buildID string) (*Cluster, error) {
	buildURL, version, err := builds.Resolve(buildID)
	if err != nil {
		return nil, err
	}
	return c.WithKubernetesBuildInfo(buildID, buildURL, version), nil
}

// WithKubernetesBuildInfo assigns the build of Kubernetes to deploy.
func (c *Cluster) WithKubernetesBuildInfo(
	buildID, buildURL, version string) *Cluster {

	if buildID != "" {
		c.Cluster.Labels[config.KubernetesBuildIDLabelName] = buildID
	}
	if buildURL != "" {
		c.Cluster.Labels[config.KubernetesBuildURLLabelName] = buildURL
	}
	for i := range c.Machines.Items {
		machine := &c.Machines.Items[i]
		if buildID != "" {
			machine.Labels[config.KubernetesBuildIDLabelName] = buildID
		}
		if buildURL != "" {
			machine.Labels[config.KubernetesBuildURLLabelName] = buildURL
		}
		machine.Spec.Versions.ControlPlane = version
		machine.Spec.Versions.Kubelet = version
	}
	return c
}

// WithName assigns the cluster's name.
func (c *Cluster) WithName(val string) (*Cluster, error) {
	if err := ValidateName(val); err != nil {
		return nil, err
	}

	parts := strings.Split(val, "-")
	machineDomain := fmt.Sprintf("%s.%s", parts[1], parts[0])
	c.Cluster.SetName(val)

	if dataDir != "" {
		c.Cluster.Labels[config.ConfigDirLabelName] = FilePath(val)
	}

	cidx, widx := 1, 1
	for i := range c.Machines.Items {
		machine := &c.Machines.Items[i]
		machine.Labels[capi.MachineClusterLabelName] = val

		role := config.MachineRoleWorker
		if i == 0 {
			role.Set(config.MachineRoleControlPlane)
		}
		machine.Labels[config.MachineRoleLabelName] = role.String()

		var machineName string
		if role.Has(config.MachineRoleControlPlane) {
			machineName = fmt.Sprintf("c%02d.%s", cidx, machineDomain)
			cidx++
		} else if role.Has(config.MachineRoleWorker) {
			machineName = fmt.Sprintf("w%02d.%s", widx, machineDomain)
			widx++
		}

		machine.SetName(machineName)
	}

	return c, nil
}

// WithNewName assigns the cluster a unique name.
func (c *Cluster) WithNewName() (*Cluster, error) {
	return c.WithName(NewName())
}

// WriteTo writes the object to the provided io.Writer.
func (c *Cluster) WriteTo(w io.Writer) (int64, error) {
	buf, err := c.MarshalText()
	if err != nil {
		return 0, err
	}
	return io.Copy(w, bytes.NewReader(buf))
}

// WriteToFile writes the object to the provided file.
func (c *Cluster) WriteToFile(path string) (int64, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return c.WriteTo(f)
}

// WriteToDisk writes the object to a file named "sk8.conf" in the directory
// specified by the cluster label config.ConfigDirLabelName.
func (c *Cluster) WriteToDisk() error {
	if c == nil {
		return nil
	}
	if len(c.Cluster.Labels) == 0 {
		return nil
	}
	confDir := c.Cluster.Labels[config.ConfigDirLabelName]
	if confDir == "" {
		return nil
	}
	os.MkdirAll(confDir, 0755)

	confFile := path.Join(confDir, "sk8.conf")
	if _, err := c.WriteToFile(confFile); err != nil {
		return err
	}

	return nil
}
