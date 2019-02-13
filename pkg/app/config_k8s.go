package app // import "vmw.io/sk8/app"

import (
	"context"
	"fmt"
	"math/rand"
	"os"
)

const (
	defaultK8sVersion = "release/stable"
)

// KubernetesConfig is the data used to configure Kubernetes.
type KubernetesConfig struct {
	// Versions of the components used when standing up the cluster.
	Version string `json:"version,omitempty"`

	// CloudProvider is the cloud provider configuration.
	CloudProvider CloudProviderConfig `json:"cloud-provider,omitempty"`

	// KubernetesLogLevel is the default log level for all of the Kubernetes
	// components. Valid log levels are 1-10. The default log level is 2.
	LogLevel                  uint8 `json:"log-level,omitempty"`
	APIServerLogLevel         uint8 `json:"api-server-log-level,omitempty"`
	SchedulerLogLevel         uint8 `json:"scheduler-log-level,omitempty"`
	ControllerManagerLogLevel uint8 `json:"controller-manager-log-level,omitempty"`
	KubeletLogLevel           uint8 `json:"kubelet-log-level,omitempty"`
	KubeProxyLogLevel         uint8 `json:"kube-proxy-log-level,omitempty"`

	// EncryptionKey is a 32 character string used by Kubernetes
	// to encrypt data. Random data is recommended unless attempting to
	// access data encrypted previously with a known key. If the value
	// is not set then a random string will be used.
	EncryptionKey string `json:"encryption-key,omitempty"`
}

func (c *KubernetesConfig) readEnv(ctx context.Context) error {
	if c.Version == "" {
		if c.Version = os.Getenv("SK8_K8S_VERSION"); c.Version == "" {
			if c.Version = os.Getenv("SK8_KUBE_VERSION"); c.Version == "" {
				c.Version = os.Getenv("SK8_KUBERNETES_VERSION")
			}
		}
	}
	if c.EncryptionKey == "" {
		c.EncryptionKey = os.Getenv("SK8_K8S_ENCRYPTION_KEY")
	}
	return c.CloudProvider.readEnv(ctx)
}

func (c *KubernetesConfig) validate(ctx context.Context) error {
	return c.CloudProvider.validate(ctx)
}

func (c *KubernetesConfig) setDefaults(ctx context.Context, cfg Config) error {
	if c.EncryptionKey == "" {
		buf := make([]byte, 16)
		if _, err := rand.Read(buf); err != nil {
			return err
		}
		c.EncryptionKey = fmt.Sprintf("%X", buf)
	}
	if c.Version == "" {
		c.Version = defaultK8sVersion
	}
	return c.CloudProvider.setDefaults(ctx, cfg)
}

func (c *KubernetesConfig) setEnv(
	ctx context.Context, env map[string]string) error {

	env["ENCRYPTION_KEY"] = c.EncryptionKey
	env["K8S_VERSION"] = c.Version
	return c.CloudProvider.setEnv(ctx, env)
}
