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
	"github.com/pkg/errors"
)

func (a actuator) ovaEnsure(ctx *reqctx) error {
	// Make sure the OVA is uploaded to the vSphere endpoint's
	// content library.
	if ctx.csta.OVAID != "" {
		return nil
	}
	ovaID, err := ctx.vcli.EnsureLibraryOVA(ctx, ctx.ccfg.OVA)
	if err != nil {
		return errors.Wrap(err, "error ensuring library OVA")
	}
	if ovaID == "" {
		return errors.New("OVA ID is missing")
	}
	ctx.csta.OVAID = ovaID
	return nil
}
