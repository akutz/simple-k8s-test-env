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

package encoding

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/yaml"

	sk8_v1alpha0 "vmware.io/sk8/pkg/config"
	vsphere_v1alpha0 "vmware.io/sk8/pkg/provider/vsphere/config"
)

var (
	// Scheme is the runtime.Scheme to which all the CAPI and sk8 API versions
	// and types are registered.
	Scheme = runtime.NewScheme()

	// Codecs provides access to encoding and decoding for the scheme.
	Codecs = serializer.NewCodecFactory(Scheme)

	// Decoder is a UniversalDecoder built from Codecs.
	Decoder = Codecs.UniversalDecoder()

	// ErrNilRawExtension occurs when a RawExtension object is nil or its
	// embedded data is nil.
	ErrNilRawExtension = errors.New("RawExtension is nil")
)

func init() {
	AddToScheme(Scheme)
}

// AddToScheme builds the scheme using all known CAPI and sk8 API versions
// and types.
func AddToScheme(scheme *runtime.Scheme) {
	utilruntime.Must(capi.AddToScheme(scheme))
	utilruntime.Must(sk8_v1alpha0.AddToScheme(scheme))
	utilruntime.Must(vsphere_v1alpha0.AddToScheme(scheme))
}

// Decode returns a new API object unmarshalled from the given data.
func Decode(data []byte) (runtime.Object, error) {
	var typeMeta runtime.TypeMeta
	if err := yaml.Unmarshal(data, &typeMeta); err != nil {
		return nil, errors.Wrap(err, "error discovering typeMeta info")
	}

	gvk := typeMeta.GroupVersionKind()
	obj, err := New(gvk)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating object %v", gvk)
	}

	if err := yaml.Unmarshal(data, obj); err != nil {
		return nil, errors.Wrapf(err, "error decoding data into object %v", gvk)
	}

	Scheme.Default(obj)
	return obj, nil
}

// DecodeInto decodes the provided data into the given obj.
func DecodeInto(data []byte, obj runtime.Object) error {
	_, _, err := Decoder.Decode(data, nil, obj)
	if err != nil {
		return err
	}
	Scheme.Default(obj)
	return nil
}

// New returns a new object for the provided group, version, and kind.
func New(gvk schema.GroupVersionKind) (runtime.Object, error) {
	obj, err := Scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return obj, nil
}

// FromRaw returns a new object from the runtime API object raw data.
func FromRaw(raw *runtime.RawExtension) (runtime.Object, error) {
	if raw == nil {
		return nil, ErrNilRawExtension
	}

	// If raw.Object is already set then just return it.
	if raw.Object != nil {
		return raw.Object, nil
	}

	// Otherwise ensure the raw data is set.
	if len(raw.Raw) == 0 {
		return nil, ErrNilRawExtension
	}

	// Decode the raw data into a new runtime.Object. This will fail if the
	// raw data is not registered with the current scheme.
	obj, err := Decode(raw.Raw)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding from raw extension")
	}

	// Replace the raw data with the decoded object.
	raw.Object = obj
	raw.Raw = nil

	return raw.Object, nil
}
