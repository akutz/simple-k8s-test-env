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

package net

import (
	"context"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

// GetConfigSpec gets a config spec for a network device.
func GetConfigSpec(
	ctx context.Context,
	netw object.NetworkReference,
	devs object.VirtualDeviceList) (types.BaseVirtualDeviceConfigSpec, error) {

	// Prepare virtual device config spec for network card.
	op := types.VirtualDeviceConfigSpecOperationAdd
	card, err := GetDevice(ctx, netw)
	if err != nil {
		return nil, err
	}

	// Search for the first network card of the source and update
	// the config spec's network device if necessary.
	for _, dev := range devs {
		if _, ok := dev.(types.BaseVirtualEthernetCard); ok {
			op = types.VirtualDeviceConfigSpecOperationEdit
			UpdateDevice(ctx, dev, card)
			card = dev
			break
		}
	}
	return &types.VirtualDeviceConfigSpec{
		Operation: op,
		Device:    card,
	}, nil
}

// GetDevice gets the network device for the given object reference.
func GetDevice(
	ctx context.Context,
	ref object.NetworkReference) (types.BaseVirtualDevice, error) {

	backing, err := ref.EthernetCardBackingInfo(ctx)
	if err != nil {
		return nil, err
	}
	dev, err := object.EthernetCardTypes().CreateEthernetCard("e1000", backing)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

// NewDevice returns a new network device spec.
func NewDevice(
	ctx context.Context,
	ref object.NetworkReference) (types.BaseVirtualDeviceConfigSpec, error) {

	dev, err := GetDevice(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &types.VirtualDeviceConfigSpec{
		Operation: types.VirtualDeviceConfigSpecOperationAdd,
		Device:    dev,
	}, nil
}

// UpdateDevice updates the device "dst" with information from the device "src".
func UpdateDevice(
	ctx context.Context,
	dst types.BaseVirtualDevice,
	src types.BaseVirtualDevice) {

	current := dst.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
	changed := src.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
	current.Backing = changed.Backing
	if changed.MacAddress != "" {
		current.MacAddress = changed.MacAddress
	}
	if changed.AddressType != "" {
		current.AddressType = changed.AddressType
	}
}
