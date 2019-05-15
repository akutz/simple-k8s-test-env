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

package kubeadm

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
)

// ConfigData is the information required to get the kubeadm config data.
type ConfigData struct {
	ClusterName          string
	KubernetesVersion    string
	ControlPlaneEndpoint string
	APIAdvertiseAddress  string
	APIBindPort          int32
	APIServerCertSANs    []string
	DNSDomain            string
	PodSubnet            string
	ServiceSubnet        string
	CloudProvider        string
	CloudConfig          string
}

// Config returns a kubeadm config generated from config data for a given
// version of Kubernetes.
func Config(data ConfigData) ([]byte, error) {
	ver, err := version.ParseGeneric(data.KubernetesVersion)
	if err != nil {
		return nil, err
	}

	// assume the latest API version, then fallback if the k8s version is too low
	templateSource := ConfigTemplateBetaV1
	if ver.LessThan(version.MustParseSemantic("v1.12.0")) {
		templateSource = ConfigTemplateAlphaV2
	} else if ver.LessThan(version.MustParseSemantic("v1.13.0")) {
		templateSource = ConfigTemplateAlphaV3
	}

	t, err := template.New("t").Parse(templateSource)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config template")
	}

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		return nil, errors.Wrap(err, "error executing config template")
	}

	return buf.Bytes(), nil
}
