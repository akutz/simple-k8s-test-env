package app_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"vmw.io/sk8/app"
	"vmw.io/sk8/config"
)

func TestValidateConfig(t *testing.T) {
	os.Setenv("SK8_DOMAIN_ID_RAND_SEED", "0")

	ctx := context.Background()

	cfg := config.Config{
		Nodes: []config.NodeConfig{
			config.NodeConfig{
				Type: config.ControlPlaneWorkerNode,
			},
			/*config.NodeConfig{
				Type: config.ControlPlaneNode,
			},
			config.NodeConfig{
				Type: config.WorkerNode,
			},*/
		},
		Users: []config.UserConfig{
			config.UserConfig{
				Name:         "akutz",
				SSHPublicKey: `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDE0c5FczvcGSh/tG4iw+Fhfi/O5/EvUM/96js65tly4++YTXK1d9jcznPS5ruDlbIZ30oveCBd3kT8LLVFwzh6hepYTf0YmCTpF4eDunyqmpCXDvVscQYRXyasEm5olGmVe05RrCJSeSShAeptv4ueIn40kZKOghinGWLDSZG4+FFfgrmcMCpx5YSCtX2gvnEYZJr0czt4rxOZuuP7PkJKgC/mt2PcPjooeX00vAj81jjU2f3XKrjjz2u2+KIt9eba+vOQ6HiC8c2IzRkUAJ5i1atLy8RIbejo23+0P4N2jjk17QySFOVHwPBDTYb0/0M/4ideeU74EN/CgVsvO6JrLsPBR4dojkV5qNbMNxIVv5cUwIy2ThlLgqpNCeFIDLCWNZEFKlEuNeSQ2mPtIO7ETxEL2Cz5y/7AIuildzYMc6wi2bofRC8HmQ7rMXRWdwLKWsR0L7SKjHblIwarxOGqLnUI+k2E71YoP7SZSlxaKi17pqkr0OMCF+kKqvcvHAQuwGqyumTEWOlH6TCx1dSPrW+pVCZSHSJtSTfDW2uzL6y8k10MT06+pVunSrWo5LHAXcS91htHV1M1UrH/tZKSpjYtjMb5+RonfhaFRNzvj7cCE1f3Kp8UVqAdcGBTtReoE8eRUT63qIxjw03a7VwAyB2w+9cu1R9/vAo8SBeRqw== sakutz@gmail.com`,
			},
		},
	}

	if err := app.ValidateConfig(ctx, &cfg); err != nil {
		t.Error(err)
	}

	enc := json.NewEncoder(&testOutputWriter{t})
	enc.SetIndent("", "  ")
	enc.Encode(cfg)

}
