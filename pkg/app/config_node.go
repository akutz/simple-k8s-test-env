package app // import "vmw.io/sk8/app"

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/vmware/govmomi/object"
)

const (
	defaultCores          uint8  = 1
	defaultCoresPerSocket uint8  = 1
	defaultDiskGiB        uint32 = 100
	defaultMemoryMiB      uint32 = 4096

	discoURIFormat = "https://discovery.etcd.io/new?size=%d"
)

// NodeConfig describes the members nodes in the cluster.
type NodeConfig struct {
	NodeConfigDefaults
	ControlPlaneDefaults NodeConfigDefaults `json:"default-control-plane,omitempty"`
	WorkerDefaults       NodeConfigDefaults `json:"default-worker,omitempty"`
	Nodes                []SingleNodeConfig `json:"nodes,omitempty"`

	discoURI       string
	numControllers uint16
}

// NodeType is the type for the enumeration that defines the types of
// nodes that may be deployed.
type NodeType uint8

const (
	// UnknownNodeType is an invalid node type.
	UnknownNodeType NodeType = iota

	// ControlPlaneNode is a member of the control plane on which
	// workloads may not be scheduled.
	ControlPlaneNode

	// WorkerNode is a node on which workloads may be scheduled.
	WorkerNode

	// ControlPlaneWorkerNode is a member of the control plane on which
	// workloads may be scheduled.
	ControlPlaneWorkerNode
)

// ParseNodeType returns a node type value.
func ParseNodeType(s string) NodeType {
	i, _ := strconv.ParseUint(s, 10, 8)
	switch NodeType(uint8(i)) {
	case ControlPlaneWorkerNode:
		return ControlPlaneWorkerNode
	case ControlPlaneNode:
		return ControlPlaneNode
	case WorkerNode:
		return WorkerNode
	default:
		return UnknownNodeType
	}
}

// NodeConfigDefaults is the default configuration for a node.
type NodeConfigDefaults struct {
	DefaultType           NodeType `json:"default-type,omitempty"`
	DefaultCores          uint8    `json:"default-cores,omitempty"`
	DefaultCoresPerSocket uint8    `json:"default-cores-per-socket,omitempty"`
	DefaultMemoryMiB      uint32   `json:"default-mem-mib,omitempty"`
	DefaultDiskGiB        uint32   `json:"default-disk-gib,omitempty"`
}

// SingleNodeConfig is the configuration information for a cluster node.
type SingleNodeConfig struct {
	// Type is the type of node.
	Type NodeType `json:"type,omitempty"`

	// Cores is the number of CPU cores for the node.
	Cores uint8 `json:"cores,omitempty"`

	// CoresPerSocket is the number of CPU cores per socket.
	CoresPerSocket uint8 `json:"cores-per-socket,omitempty"`

	// MemoryMiB is the amount of RAM (MiB) for the node.
	MemoryMiB uint32 `json:"mem-mib,omitempty"`

	// DiskGiB is the amount of disk space (GiB) for the node.
	DiskGiB uint32 `json:"disk-gib,omitempty"`

	// The following values are set when a cluster is turned up and are
	// ignored while bringing up a cluster.
	Name string `json:"name,omitempty"`
	FQDN string `json:"fqdn,omitempty"`
	UUID string `json:"uuid,omitempty"`

	hostName string
	vm       *object.VirtualMachine
}

func (c *NodeConfig) readEnv(ctx context.Context) error {
	if c.DefaultType == UnknownNodeType {
		c.DefaultType = ParseNodeType(os.Getenv("SK8_NODE_DEFAULT_TYPE"))
	}
	if err := c.NodeConfigDefaults.readEnv(
		ctx,
		"SK8_NODE_DEFAULT_CORES", "SK8_NODE_DEFAULT_CORES_PER_SOCKET",
		"SK8_NODE_DEFAULT_DISK_GIB", "SK8_NODE_DEFAULT_MEM_MIB"); err != nil {
		return err
	}
	if err := c.ControlPlaneDefaults.readEnv(
		ctx,
		"SK8_NODE_DEFAULT_CONTROLLER_CORES",
		"SK8_NODE_DEFAULT_CONTROLLER_CORES_PER_SOCKET",
		"SK8_NODE_DEFAULT_CONTROLLER_DISK_GIB",
		"SK8_NODE_DEFAULT_CONTROLLER_MEM_MIB"); err != nil {
		return err
	}
	if err := c.WorkerDefaults.readEnv(
		ctx,
		"SK8_NODE_DEFAULT_WORKER_CORES",
		"SK8_NODE_DEFAULT_WORKER_CORES_PER_SOCKET",
		"SK8_NODE_DEFAULT_WORKER_DISK_GIB",
		"SK8_NODE_DEFAULT_WORKER_MEM_MIB"); err != nil {
		return err
	}

	return nil
}

func (c *NodeConfig) validate(ctx context.Context) error {
	return nil
}

func (c *NodeConfig) setDefaults(ctx context.Context, cfg Config) error {
	c.discoURI = ""
	c.numControllers = 0

	if c.DefaultType == UnknownNodeType {
		c.DefaultType = ControlPlaneWorkerNode
	}
	if err := c.NodeConfigDefaults.setDefaults(
		ctx, NodeConfigDefaults{
			DefaultCores:          defaultCores,
			DefaultCoresPerSocket: defaultCoresPerSocket,
			DefaultDiskGiB:        defaultDiskGiB,
			DefaultMemoryMiB:      defaultMemoryMiB,
		}); err != nil {
		return err
	}
	if err := c.ControlPlaneDefaults.setDefaults(
		ctx, c.NodeConfigDefaults); err != nil {
		return err
	}
	if err := c.WorkerDefaults.setDefaults(
		ctx, c.NodeConfigDefaults); err != nil {
		return err
	}
	if len(c.Nodes) == 0 {
		c.Nodes = append(c.Nodes, SingleNodeConfig{})
	}

	var workerNodeIndex int
	for i := range c.Nodes {
		n := &c.Nodes[i]
		if n.Type == UnknownNodeType {
			n.Type = ControlPlaneWorkerNode
		}
		switch n.Type {
		case ControlPlaneNode:
			c.numControllers++
			n.hostName = fmt.Sprintf("c%02d", c.numControllers)
			if err := n.setDefaults(ctx, c.ControlPlaneDefaults); err != nil {
				return err
			}
		case ControlPlaneWorkerNode, WorkerNode:
			if n.Type == ControlPlaneWorkerNode {
				c.numControllers++
				n.hostName = fmt.Sprintf("c%02d", c.numControllers)
			} else {
				workerNodeIndex++
				n.hostName = fmt.Sprintf("w%02d", workerNodeIndex)
			}
			if err := n.setDefaults(ctx, c.WorkerDefaults); err != nil {
				return err
			}
		}
		n.FQDN = fmt.Sprintf("%s.%s", n.hostName, cfg.Network.DomainFQDN)
		n.Name = fmt.Sprintf(
			"%s.%s", n.hostName, strings.Split(cfg.Network.DomainFQDN, ".")[0])
	}
	if len(c.Nodes) > 1 {
		discoURL := fmt.Sprintf(discoURIFormat, c.numControllers)
		resp, err := http.Get(discoURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		c.discoURI = string(buf)
	}

	return nil
}

func (c *NodeConfig) setEnv(ctx context.Context, env map[string]string) error {
	env["NUM_NODES"] = strconv.Itoa(len(c.Nodes))
	env["NUM_CONTROLLERS"] = strconv.Itoa(int(c.numControllers))
	env["ETCD_DISCOVERY"] = c.discoURI
	return nil
}

func (c *NodeConfigDefaults) readEnv(
	ctx context.Context,
	coresName, coresPerSocketName, diskGiBName, memMiBName string) error {

	if c.DefaultCores == 0 {
		i, _ := strconv.ParseUint(os.Getenv(coresName), 10, 8)
		c.DefaultCores = uint8(i)
	}
	if c.DefaultCoresPerSocket == 0 {
		i, _ := strconv.ParseUint(os.Getenv(coresPerSocketName), 10, 8)
		c.DefaultCoresPerSocket = uint8(i)
	}
	if c.DefaultDiskGiB == 0 {
		i, _ := strconv.ParseUint(os.Getenv(diskGiBName), 10, 32)
		c.DefaultDiskGiB = uint32(i)
	}
	if c.DefaultMemoryMiB == 0 {
		i, _ := strconv.ParseUint(os.Getenv(memMiBName), 10, 32)
		c.DefaultMemoryMiB = uint32(i)
	}
	return nil
}

func (c *NodeConfigDefaults) setDefaults(
	ctx context.Context, src NodeConfigDefaults) error {

	if c.DefaultCores == 0 {
		c.DefaultCores = src.DefaultCores
	}
	if c.DefaultCoresPerSocket == 0 {
		c.DefaultCoresPerSocket = src.DefaultCoresPerSocket
	}
	if c.DefaultDiskGiB == 0 {
		c.DefaultDiskGiB = src.DefaultDiskGiB
	}
	if c.DefaultMemoryMiB == 0 {
		c.DefaultMemoryMiB = src.DefaultMemoryMiB
	}
	return nil
}

func (c *SingleNodeConfig) setDefaults(
	ctx context.Context, src NodeConfigDefaults) error {

	if c.Cores == 0 {
		c.Cores = src.DefaultCores
	}
	if c.CoresPerSocket == 0 {
		c.CoresPerSocket = src.DefaultCoresPerSocket
	}
	if c.DiskGiB == 0 {
		c.DiskGiB = src.DefaultDiskGiB
	}
	if c.MemoryMiB == 0 {
		c.MemoryMiB = src.DefaultMemoryMiB
	}
	return nil
}
