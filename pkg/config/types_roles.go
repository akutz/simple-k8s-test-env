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

package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MachineRole is one or more roles supported by a kubelet installed on
// a machine.
type MachineRole int32

const (
	// MachineRoleControlPlane indicates a machine is a member of the
	// Kubernetes control plane.
	MachineRoleControlPlane MachineRole = 1 << iota

	// MachineRoleWorker indicates a machine is a worker node.
	MachineRoleWorker
)

// Clear removes a role.
func (r *MachineRole) Clear(role MachineRole) {
	*r = *r &^ role
}

// Has returns a flag indicating whether or not the provided role is set.
func (r MachineRole) Has(role MachineRole) bool {
	return r&role != 0
}

// Set adds a role.
func (r *MachineRole) Set(role MachineRole) {
	*r = *r | role
}

// Format ensures %d and %v format the MachineRole as its numeric mask value.
func (r MachineRole) Format(f fmt.State, c rune) {
	switch c {
	case 'v', 'd':
		fmt.Fprint(f, int32(r))
	default:
		fmt.Fprint(f, r.String())
	}
}

// String returns a CSV string of friendly role names.
func (r MachineRole) String() string {
	if r.Has(MachineRoleControlPlane) && r.Has(MachineRoleWorker) {
		return "control-plane,worker"
	} else if r.Has(MachineRoleControlPlane) {
		return "control-plane"
	} else if r.Has(MachineRoleWorker) {
		return "worker"
	}
	return ""
}

// MarshalJSON encodes the MachineRole as a JSON array of strings.
func (r MachineRole) MarshalJSON() ([]byte, error) {
	var roles []string
	if r.Has(MachineRoleControlPlane) {
		roles = append(roles, "control-plane")
	}
	if r.Has(MachineRoleWorker) {
		roles = append(roles, "worker")
	}
	return json.Marshal(roles)
}

// MarshalText encodes the MachineRole as a CSV string of friendly role names.
func (r MachineRole) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// UnmarshalJSON decodes the MachineRole from a JSON array of strings.
func (r *MachineRole) UnmarshalJSON(data []byte) error {
	var roles []string
	if err := json.Unmarshal(data, &roles); err != nil {
		return err
	}
	for _, s := range roles {
		switch s {
		case "control-plane":
			r.Set(MachineRoleControlPlane)
		case "worker":
			r.Set(MachineRoleWorker)
		}
	}
	return nil
}

// UnmarshalText decodes the MachineRole using ParseMachineRole.
func (r *MachineRole) UnmarshalText(data []byte) error {
	parseMachineRoleStrings(r, strings.Split(string(data), ","))
	return nil
}

// ParseMachineRole returns a MachineRole mask from a CSV string of friendly
// role names.
func ParseMachineRole(s string) MachineRole {
	var r MachineRole
	parseMachineRoleStrings(&r, strings.Split(s, ","))
	return r
}

func parseMachineRoleStrings(r *MachineRole, roles []string) {
	for _, s := range roles {
		switch s {
		case "control-plane":
			r.Set(MachineRoleControlPlane)
		case "worker":
			r.Set(MachineRoleWorker)
		}
	}
}
