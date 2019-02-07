package config_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/vmware/sk8/pkg/config"
)

func TestConfigMarshalJSON(t *testing.T) {
	c := config.Config{
		Nodes: []config.NodeConfig{
			config.NodeConfig{
				Type:           config.ControlPlaneWorkerNode,
				Cores:          16,
				CoresPerSocket: 8,
				MemoryMiB:      65536,
				DiskGiB:        100,
			},
		},
	}
	var w io.Writer = ioutil.Discard
	if testing.Verbose() {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		t.Fatal(err)
	}
}

func TestConfigUnmarshalJSON(t *testing.T) {
	data := `{
  "nodes": [
    {
      "type": 2,
      "cores": 16,
      "cores-per-socket": 8,
      "mem": 65536,
      "disk": 100
    }
  ],
  "k8s": {
    "log-level": 2,
    "cloud-provider": {
      "type": 2
    }
  }
}`
	c := config.Config{}
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		t.Fatal(err)
	}
	if len(c.Nodes) != 1 {
		t.Fatal("len(c.Nodes) != 1")
	}
	if c.Nodes[0].Type != config.WorkerNode {
		t.Fatal("c.Nodes[0].Type != config.WorkerNode")
	}
	if c.Nodes[0].Cores != 16 {
		t.Fatal("c.Nodes[0].Cores != 16")
	}
	if c.Nodes[0].CoresPerSocket != 8 {
		t.Fatal("c.Nodes[0].CoresPerSocket != 8")
	}
	if c.Nodes[0].MemoryMiB != 65536 {
		t.Fatal("c.Nodes[0].MemoryMiB != 65536")
	}
	if c.Nodes[0].DiskGiB != 100 {
		t.Fatal("c.Nodes[0].DiskGiB != 100")
	}
	if c.K8s.LogLevel != 2 {
		t.Fatal("c.K8s.LogLevel != 2")
	}
	if c.K8s.CloudProvider.Type != config.ExternalCloudProvider {
		t.Fatal("c.K8s.CloudProvider.Type != config.ExternalCloudProvider")
	}

}
