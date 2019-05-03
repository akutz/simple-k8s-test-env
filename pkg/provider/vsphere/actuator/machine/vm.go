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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/vapi/vcenter"
	"github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/net/lvs"
	vcloudinit "vmware.io/sk8/pkg/provider/vsphere/cloudinit"
	vdisk "vmware.io/sk8/pkg/provider/vsphere/govmomi/disk"
	vnet "vmware.io/sk8/pkg/provider/vsphere/govmomi/net"
)

func (a actuator) vmEnsure(ctx *reqctx) error {

	// Find or create the VM.
	if vm, err := ctx.vcli.Finder.VirtualMachine(
		ctx, ctx.machine.Name); err == nil && vm != nil {
		log.WithField("vm", ctx.machine.Name).Debug("already exists")
		ctx.vm = vm
	} else {
		deploySpec := vcenter.Deploy{
			Target: vcenter.Target{
				ResourcePoolID: ctx.vcli.ResourcePool.Reference().Value,
				FolderID:       ctx.vcli.Folder.Reference().Value,
			},
			DeploymentSpec: vcenter.DeploymentSpec{
				Name:               ctx.machine.Name,
				AcceptAllEULA:      true,
				DefaultDatastoreID: ctx.vcli.Datastore.Reference().Value,
			},
		}

		// Deploy the VM.
		vm, err := ctx.vcli.DeployOVF(ctx, ctx.csta.OVAID, deploySpec)
		if err != nil {
			return errors.Wrapf(
				err, "error deploying ovf to %q", ctx.machine.Name)
		}

		devChanges := []types.BaseVirtualDeviceConfigSpec{}

		// Get the devices for the newly deployed VM.
		devices, err := vm.Device(ctx)
		if err != nil {
			return errors.Wrapf(
				err, "error getting devices for %q", ctx.machine.Name)
		}

		// If the external access method is LVS then add another
		// network interface.
		for i, iface := range ctx.mcfg.Network.Interfaces {
			netName := iface.Network
			netRef := ctx.vcli.Networks[netName]

			if i == 0 {
				// Search for the first network card of the source and update
				// the config spec's network device if necessary.
				spec, err := vnet.GetConfigSpec(ctx, netRef, devices)
				if err != nil {
					return errors.Wrapf(
						err,
						"error getting network device for %q on node %q",
						netName, ctx.machine.Name)
				}
				devChanges = append(devChanges, spec)
			} else {
				spec, err := vnet.NewDevice(ctx, netRef)
				if err != nil {
					return errors.Wrapf(
						err,
						"error getting network device for %q on node %q",
						netName, ctx.machine.Name)

				}
				devChanges = append(devChanges, spec)
			}
		}

		// Resize the VM's disk.
		diskResizeSpec, err := vdisk.GetConfigSpec(ctx, 100, devices)
		if err != nil {
			return errors.Wrapf(
				err,
				"error getting disk device config spec for %q",
				ctx.machine.Name)
		}
		devChanges = append(devChanges, diskResizeSpec)

		// Get the extra config data for the VM.
		extraConfig, err := vcloudinit.New(ctx, ctx.cluster, ctx.machine)
		if err != nil {
			return errors.Wrapf(
				err,
				"error getting extra config %q",
				ctx.machine.Name)
		}

		vmConfigSpec := types.VirtualMachineConfigSpec{
			Name:              ctx.machine.Name,
			MemoryMB:          int64(32768),
			NumCPUs:           int32(8),
			NumCoresPerSocket: int32(4),
			DeviceChange:      devChanges,
			ExtraConfig:       extraConfig,
		}

		// Reconfigure the VM.
		{
			task, err := vm.Reconfigure(ctx, vmConfigSpec)
			if err != nil {
				return errors.Wrapf(
					err,
					"error reconfiguring %q",
					ctx.machine.Name)
			}
			result, err := task.WaitForResult(ctx, nil)
			if err != nil {
				return errors.Wrapf(
					err,
					"error waiting for reconfiguration of %q",
					ctx.machine.Name)
			}
			if result.Error != nil && result.Error.LocalizedMessage != "" {
				return errors.Wrapf(
					errors.New(result.Error.LocalizedMessage),
					"localized reconfiguration error for %q",
					ctx.machine.Name)
			}
		}

		// Store the VM's UUID with the node.
		log.WithField("vm", ctx.machine.Name).Debug("created")
		ctx.vm = vm
	}

	return a.vmEnsurePoweredOn(ctx)
}

func (a actuator) vmEnsurePoweredOn(ctx *reqctx) error {
	powerState, _ := ctx.vm.PowerState(ctx)
	if powerState != types.VirtualMachinePowerStatePoweredOn {
		log.WithField("vm", ctx.machine.Name).Debug("powering on")
		task, err := ctx.vm.PowerOn(ctx)
		if err != nil {
			return errors.Wrapf(
				err, "error powering on %q", ctx.machine.Name)
		}
		log.WithField("vm", ctx.machine.Name).Debug(
			"waiting for power on event to complete")
		result, err := task.WaitForResult(ctx, nil)
		if err != nil {
			return errors.Wrapf(
				err,
				"error waiting for %q to be powered on",
				ctx.machine.Name)
		}
		if result.Error != nil && result.Error.LocalizedMessage != "" {
			return errors.Wrapf(
				errors.New(result.Error.LocalizedMessage),
				"localized power on event error for %q",
				ctx.machine.Name)
		}
	}

	log.WithField("vm", ctx.machine.Name).Debug("powered on")
	return a.vmEnsureNetworkReady(ctx)
}

func (a actuator) vmEnsureNetworkReady(ctx *reqctx) error {
	log.WithField(
		"vm",
		ctx.machine.Name).Debug("waiting for IP address(es)")

	waitForDevice := func(
		devName string,
		addrType corev1.NodeAddressType) error {

		devices, err := ctx.vm.WaitForNetIP(ctx, true, devName)
		if err != nil {
			return errors.Wrapf(
				err, "error waiting for ip addr for %q %q",
				ctx.machine.Name,
				devName)
		}
		for macAddr, dev := range devices {
			for _, ipAddr := range dev {
				log.WithFields(log.Fields{
					"vm":      ctx.machine.Name,
					"device":  devName,
					"macAddr": macAddr,
					"ipAddr":  ipAddr,
				}).Debug("has IP address")
				ctx.machine.Status.Addresses = append(
					ctx.machine.Status.Addresses,
					corev1.NodeAddress{
						Type:    addrType,
						Address: ipAddr,
					},
				)
			}
		}
		return nil
	}

	var addrType corev1.NodeAddressType
	switch ctx.ccfg.NAT.Object.(type) {
	case *config.LinuxVirtualSwitchConfig:
		addrType = lvs.NodeIP
	default:
		addrType = corev1.NodeInternalIP
	}

	if err := waitForDevice("ethernet-0", addrType); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"vm":    ctx.machine.Name,
		"addrs": fmt.Sprintf("%+v", ctx.machine.Status.Addresses),
	}).Debug("waited for IP address(es)")

	return nil
}

func (a actuator) vmDestroy(ctx *reqctx) error {
	if vm, err := ctx.vcli.Finder.VirtualMachine(
		ctx, ctx.machine.Name); err == nil {
		if state, err := vm.PowerState(ctx); err == nil {
			if state != types.VirtualMachinePowerStatePoweredOff {
				if _, err := vm.PowerOff(ctx); err == nil {
					vm.WaitForPowerState(
						ctx,
						types.VirtualMachinePowerStatePoweredOff)
				}
			}
		}
		vm.Destroy(ctx)
	}
	return nil
}

func (a actuator) vmExists(ctx *reqctx) (bool, error) {
	if _, err := ctx.vcli.Finder.VirtualMachine(
		ctx, ctx.machine.Name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
