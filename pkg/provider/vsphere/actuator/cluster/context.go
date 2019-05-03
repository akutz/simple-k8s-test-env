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
	"context"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	vconfig "vmware.io/sk8/pkg/provider/vsphere/config"
	vclient "vmware.io/sk8/pkg/provider/vsphere/govmomi/client"
	vutil "vmware.io/sk8/pkg/provider/vsphere/util"
)

type reqctx struct {
	context.Context
	cluster *capi.Cluster
	ccfg    *vconfig.ClusterProviderConfig
	csta    *vconfig.ClusterStatus
	vcli    *vclient.Client
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
	cluster *capi.Cluster) (*reqctx, error) {

	ctx := &reqctx{
		Context: parent,
		cluster: cluster,
	}

	// Ensure the status object is present and ready.
	ctx.csta = &vconfig.ClusterStatus{
		ControlPlaneOnline: make(chan struct{}),
		ControlPlaneCanOwn: make(chan struct{}, 1),
	}
	ctx.csta.ControlPlaneCanOwn <- struct{}{}
	if ps := cluster.Status.ProviderStatus; ps == nil {
		ps = &runtime.RawExtension{
			Object: ctx.csta,
		}
		cluster.Status.ProviderStatus = ps
	} else if csta2, ok := ps.Object.(*vconfig.ClusterStatus); ok {
		ctx.csta = csta2
	} else {
		ps.Object = ctx.csta
	}

	ctx.ccfg = vutil.GetClusterProviderConfig(cluster)

	{
		c, err := vclient.New(ctx, *ctx.ccfg)
		if err != nil {
			return nil, errors.Wrap(err, "error creating govmomi client")
		}
		ctx.vcli = c
	}

	return ctx, nil
}
