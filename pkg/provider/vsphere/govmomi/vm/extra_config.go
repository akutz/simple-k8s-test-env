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

package vm

import (
	"github.com/vmware/govmomi/vim25/types"

	"vmware.io/sk8/pkg/util"
)

// ExtraConfig is data used with a VM's guestInfo RPC interface.
type ExtraConfig []types.BaseOptionValue

// SetCloudInitUserData sets the cloud init user data at the key
// "guestinfo.userdata" as a gzipped, base64-encoded string.
func (e *ExtraConfig) SetCloudInitUserData(data []byte) error {
	encData, err := util.Base64GzipBytes(data)
	if err != nil {
		return err
	}

	*e = append(*e,
		&types.OptionValue{
			Key:   "guestinfo.userdata",
			Value: encData,
		},
		&types.OptionValue{
			Key:   "guestinfo.userdata.encoding",
			Value: "gzip+base64",
		},
	)

	return nil
}

// SetCloudInitMetadata sets the cloud init user data at the key
// "guestinfo.metadata" as a gzipped, base64-encoded string.
func (e *ExtraConfig) SetCloudInitMetadata(data []byte) error {
	encData, err := util.Base64GzipBytes(data)
	if err != nil {
		return err
	}

	*e = append(*e,
		&types.OptionValue{
			Key:   "guestinfo.metadata",
			Value: encData,
		},
		&types.OptionValue{
			Key:   "guestinfo.metadata.encoding",
			Value: "gzip+base64",
		},
	)

	return nil
}
