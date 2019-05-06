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
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"vmware.io/sk8/pkg/config"
)

// PrintInfo prints information about the cluster.
func PrintInfo(w io.Writer, format string, cluster *Cluster) error {
	info, err := GetInfo(cluster)
	if err != nil {
		return err
	}

	var tplFormat string

	switch format {
	case "json":
		enc := json.NewEncoder(w)
		return enc.Encode(info)
	case "yaml":
		buf, err := yaml.Marshal(info)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, bytes.NewReader(buf)); err != nil {
			return err
		}
		return nil
	case "text":
		tplFormat = DefaultTemplate
	default:
		tplFormat = format
	}

	tpl := template.Must(template.New("t").Parse(tplFormat))
	return tpl.Execute(w, info)
}

// GetInfo returns a summary of a cluster's information.
func GetInfo(c *Cluster) (*Info, error) {
	info := &Info{
		Name:     c.Cluster.Name,
		Machines: make([]MachineInfo, len(c.Machines.Items)),
		Program:  os.Args[0],
	}
	if t := c.Cluster.CreationTimestamp; !t.IsZero() {
		info.Created = &t
	}
	if t := c.Cluster.DeletionTimestamp; t != nil && !t.IsZero() {
		info.Deleted = t
	}
	if labels := c.Cluster.Labels; len(labels) > 0 {
		if dir := labels[config.ConfigDirLabelName]; dir != "" {
			info.Kubeconfig = path.Join(dir, "kube.conf")
		}
		if ccm := labels[config.CloudProviderLabelName]; ccm != "" {
			info.CloudProvider = ccm
		}
	}

	for i := range info.Machines {
		machine := c.Machines.Items[i]
		machineInfo := &info.Machines[i]
		machineInfo.Name = machine.Name
		if t := machine.CreationTimestamp; !t.IsZero() {
			machineInfo.Created = &t
		}
		if t := machine.DeletionTimestamp; t != nil && !t.IsZero() {
			machineInfo.Deleted = t
		}
		if labels := machine.Labels; len(labels) > 0 {
			if roles := labels[config.MachineRoleLabelName]; roles != "" {
				if err := machineInfo.Roles.UnmarshalText(
					[]byte(roles)); err != nil {
					return nil, err
				}
			}
		}
		if v := machine.Spec.Versions; true {
			machineInfo.Versions.ControlPlane = v.ControlPlane
			machineInfo.Versions.Kubelet = v.Kubelet
		}
	}
	return info, nil
}

// Info is a summary of a cluster's information. This object is useful
// when emitting data as JSON, YAML, or text.
type Info struct {
	Name          string        `json:"name"`
	Created       *metav1.Time  `json:"created,omitempty"`
	Deleted       *metav1.Time  `json:"deleted,omitempty"`
	Kubeconfig    string        `json:"kubeconfig"`
	CloudProvider string        `json:"cloudProvider,omitempty"`
	Machines      []MachineInfo `json:"machines"`
	Program       string        `json:"-"`
}

// MachineInfo is a summary of a machine's information. This object is
// useful when emitting data as JSON, YAML, or text.
type MachineInfo struct {
	Name     string             `json:"name"`
	Created  *metav1.Time       `json:"created,omitempty"`
	Deleted  *metav1.Time       `json:"deleted,omitempty"`
	Roles    config.MachineRole `json:"roles"`
	Versions struct {
		ControlPlane string `json:"controlPlane,omitempty"`
		Kubelet      string `json:"kubelet"`
	} `json:"versions"`
}

// IsController is a function that can be used by a Go template to determine
// whether or not a machine is a member of the control plane.
func (m MachineInfo) IsController() bool {
	return m.Roles.Has(config.MachineRoleControlPlane)
}
