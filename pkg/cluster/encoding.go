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
	"encoding/json"
	"reflect"
)

// MarshalJSONNoEmptyVals marshals the provided object to JSON and
// strips the JSON of any keys with empty values.
func MarshalJSONNoEmptyVals(v interface{}) ([]byte, error) {
	m := map[string]interface{}{}
	{
		buf, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if json.Unmarshal(buf, &m); err != nil {
			return nil, err
		}
	}
	RemoveEmptyValues(m)
	buf, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// RemoveEmptyValues removes empty values from a map.
func RemoveEmptyValues(m map[string]interface{}) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
		} else if m2, ok := v.(map[string]interface{}); ok {
			if len(m2) == 0 {
				delete(m, k)
			} else {
				RemoveEmptyValues(m2)
				if len(m2) == 0 {
					delete(m, k)
				}
			}
		} else {
			vv := reflect.ValueOf(v)
			switch vv.Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				if vv.Len() == 0 {
					delete(m, k)
				}
			case reflect.String:
				if vv.String() == "" {
					delete(m, k)
				}
			}
		}
	}
}
