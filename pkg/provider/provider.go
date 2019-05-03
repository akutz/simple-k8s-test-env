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

package provider

import (
	"reflect"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/cluster-api/pkg/controller/cluster"
	"sigs.k8s.io/cluster-api/pkg/controller/machine"
)

var (
	registry   = map[schema.GroupKind]reflect.Type{}
	registryMu sync.RWMutex
)

// RegisterClusterActuator records a cluster.Actuator for the given group name.
func RegisterClusterActuator(group string, actuator cluster.Actuator) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[clusterRegKey(group)] = reflect.TypeOf(actuator)
}

// RegisterMachineActuator records a machine.Actuator for the given group name.
func RegisterMachineActuator(group string, actuator machine.Actuator) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[machineRegKey(group)] = reflect.TypeOf(actuator)
}

// NewClusterActuator returns a new cluster.Actuator for the given group name.
// Nil is returned if no actuator is registered for the given group name.
func NewClusterActuator(group string) cluster.Actuator {
	registryMu.RLock()
	defer registryMu.RUnlock()
	t, ok := registry[clusterRegKey(group)].(reflect.Type)
	if !ok {
		return nil
	}
	return reflect.New(t).Elem().Interface().(cluster.Actuator)
}

// NewMachineActuator returns a new machine.Actuator for the given group name.
// Nil is returned if no actuator is registered for the given group name.
func NewMachineActuator(group string) machine.Actuator {
	registryMu.RLock()
	defer registryMu.RUnlock()
	t, ok := registry[machineRegKey(group)].(reflect.Type)
	if !ok {
		return nil
	}
	return reflect.New(t).Elem().Interface().(machine.Actuator)
}

func clusterRegKey(group string) schema.GroupKind {
	return schema.GroupKind{Group: group, Kind: "ClusterActuator"}
}

func machineRegKey(group string) schema.GroupKind {
	return schema.GroupKind{Group: group, Kind: "MachineActuator"}
}
