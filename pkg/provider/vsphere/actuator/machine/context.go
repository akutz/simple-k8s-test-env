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

package machine

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/object"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/ssh"
	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
	vclient "vmware.io/sk8/pkg/provider/vsphere/govmomi/client"
	vutil "vmware.io/sk8/pkg/provider/vsphere/util"
)

type reqctx struct {
	context.Context
	cluster *capi.Cluster
	machine *capi.Machine
	dir     string
	ccfg    *vconfig.ClusterProviderConfig
	mcfg    *vconfig.MachineProviderConfig
	csta    *vconfig.ClusterStatus
	msta    *vconfig.MachineStatus
	role    config.MachineRole
	ssh     *ssh.Client
	vcli    *vclient.Client
	vm      *object.VirtualMachine
}

func (c *reqctx) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *reqctx) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *reqctx) Err() error {
	return c.Context.Err()
}

func (c *reqctx) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

func newRequestContext(
	parent context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine) (*reqctx, error) {

	ctx := &reqctx{
		Context: parent,
		cluster: cluster,
		machine: machine,
		ccfg:    vutil.GetClusterProviderConfig(cluster),
		mcfg:    vutil.GetMachineProviderConfig(machine),
		msta:    &vconfig.MachineStatus{},
		dir:     cluster.Labels[config.ConfigDirLabelName],
	}

	if ps := cluster.Status.ProviderStatus; ps != nil {
		if obj, ok := ps.Object.(*vconfig.ClusterStatus); ok {
			ctx.csta = obj
		}
	}

	if ps := machine.Status.ProviderStatus; ps == nil {
		ps = &runtime.RawExtension{
			Object: ctx.msta,
		}
	} else if msta2, ok := ps.Object.(*vconfig.MachineStatus); ok {
		ctx.msta = msta2
	} else {
		ps.Object = ctx.msta
	}

	if r, ok := machine.Labels[config.MachineRoleLabelName]; ok {
		if err := ctx.role.UnmarshalText([]byte(r)); err != nil {
			return nil, err
		}
	}

	{
		c, err := vclient.New(ctx, *ctx.ccfg)
		if err != nil {
			return nil, errors.Wrap(err, "error creating govmomi client")
		}
		if _, err := c.WithMachineProviderConfig(ctx, *ctx.mcfg); err != nil {
			return nil, errors.Wrap(
				err, "error updating vm client w/ machine provider config")
		}
		ctx.vcli = c
	}

	return ctx, nil
}
