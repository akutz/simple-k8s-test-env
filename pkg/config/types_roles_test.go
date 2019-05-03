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

package config_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"vmware.io/sk8/pkg/config"
)

func TestMachineRole(t *testing.T) {
	cases := []struct {
		TestName    string
		ExpectedObj interface{}
		ExpectError bool
		Run         func() (interface{}, error)
	}{
		{
			TestName:    "Single role",
			ExpectedObj: config.MachineRoleControlPlane,
			Run: func() (interface{}, error) {
				return config.MachineRoleControlPlane, nil
			},
		},
		{
			TestName:    "Two roles",
			ExpectedObj: config.MachineRoleControlPlane | config.MachineRoleWorker,
			Run: func() (interface{}, error) {
				return config.MachineRoleControlPlane | config.MachineRoleWorker, nil
			},
		},
		{
			TestName:    "Empty role",
			ExpectedObj: "",
			Run: func() (interface{}, error) {
				var r config.MachineRole
				return r.String(), nil
			},
		},
		{
			TestName:    "Stringer",
			ExpectedObj: "control-plane",
			Run: func() (interface{}, error) {
				return config.MachineRoleControlPlane.String(), nil
			},
		},
		{
			TestName:    "Format as string",
			ExpectedObj: "control-plane,worker",
			Run: func() (interface{}, error) {
				return fmt.Sprintf("%s",
					config.MachineRoleWorker|
						config.MachineRoleControlPlane), nil
			},
		},
		{
			TestName:    "Format as digit",
			ExpectedObj: "3",
			Run: func() (interface{}, error) {
				return fmt.Sprintf("%d",
					config.MachineRoleWorker|
						config.MachineRoleControlPlane), nil
			},
		},
		{
			TestName:    "Format as JSON",
			ExpectedObj: `["worker"]`,
			Run: func() (interface{}, error) {
				buf, _ := json.Marshal(config.MachineRoleWorker)
				return string(buf), nil
			},
		},
		{
			TestName:    "Format as default",
			ExpectedObj: "3",
			Run: func() (interface{}, error) {
				return fmt.Sprintf("%v",
					config.MachineRoleWorker|
						config.MachineRoleControlPlane), nil
			},
		},
	}
	for _, c := range cases {
		t.Run(c.TestName, func(t *testing.T) {
			actObj, err := c.Run()
			if err != nil && !c.ExpectError {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && c.ExpectError {
				t.Fatalf("unexpected lack or error")
			}
			if diff := cmp.Diff(c.ExpectedObj, actObj); diff != "" {
				t.Fatalf("diff\n\n%s\n", diff)
			}
		})
	}
}
