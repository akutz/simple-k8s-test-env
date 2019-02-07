package app // import "vmw.io/sk8/app"

import (
	"github.com/vmware/govmomi/object"
	"vmw.io/sk8/config"
)

// State is the result of an Up operation.
type State struct {
	// ClusterID is a unique ID that identifies the cluster.
	ClusterID string `json:"clusterID"`

	// Nodes are the nodes that belong to the cluster.
	Nodes []NodeInfo `json:"nodes"`
}

// NodeInfo is information about a running node.
type NodeInfo struct {
	Type     config.NodeType
	UUID     string `json:"uuid"`
	HostName string `json:"hostName"`
	HostFQDN string `json:"hostFQDN"`
	vm       *object.VirtualMachine
	vmName   string
}
