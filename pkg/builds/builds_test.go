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

package builds_test

import (
	"testing"

	"vmware.io/sk8/pkg/builds"
)

func TestResolve(t *testing.T) {
	cases := []struct {
		BuildID     string
		ExpectError bool
	}{
		{
			BuildID: "release/stable",
		},
		{
			BuildID: "release/latest.txt",
		},
		{
			BuildID: "ci/latest",
		},
		{
			BuildID: "ci/latest.txt",
		},
		{
			BuildID: "v1.12.0",
		},
		{
			BuildID: "release/latest-1.13",
		},
		{
			BuildID: "release/latest-1.14.txt",
		},
	}

	for _, c := range cases {
		t.Run(c.BuildID, func(t *testing.T) {
			uri, version, err := builds.Resolve(c.BuildID)
			if err != nil && !c.ExpectError {
				t.Fatal(err)
			}
			if err == nil && c.ExpectError {
				t.Fatalf("expected error to occur for %s", c.BuildID)
			}
			t.Logf("uri=%s version=%s", uri, version)
		})
	}
}
