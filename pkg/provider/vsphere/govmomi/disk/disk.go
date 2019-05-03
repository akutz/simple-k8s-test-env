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

package disk

import (
	"context"

	"github.com/pkg/errors"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

// GetConfigSpec gets a config spec for a disk device.
func GetConfigSpec(
	ctx context.Context,
	sizeGiB uint32,
	devs object.VirtualDeviceList) (types.BaseVirtualDeviceConfigSpec, error) {

	// Search for the first disk and update its size.
	for _, dev := range devs {
		if disk, ok := dev.(*types.VirtualDisk); ok {
			disk.CapacityInKB = int64(sizeGiB) * 1024 * 1024
			return &types.VirtualDeviceConfigSpec{
				Operation: types.VirtualDeviceConfigSpecOperationEdit,
				Device:    disk,
			}, nil
		}
	}

	return nil, errors.New("no disks found")
}
