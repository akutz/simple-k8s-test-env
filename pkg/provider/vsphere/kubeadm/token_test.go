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

package kubeadm_test

import (
	"regexp"
	"testing"

	"vmware.io/sk8/pkg/provider/vsphere/kubeadm"
)

func TestNewBootstrapToken(t *testing.T) {
	token := kubeadm.NewBootstrapToken()
	re := `[a-z0-9]{6}\.[a-z0-9]{16}`
	ok, err := regexp.MatchString(re, token)
	if err != nil {
		t.Fatalf("token=%q does not match %q: %v", token, re, err)
	}
	if !ok {
		t.Fatalf("token=%q does not match %q", token, re)
	}
}
