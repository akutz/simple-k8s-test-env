package app // import "vmw.io/sk8/app"

import (
	"context"
	"os"
	"regexp"
	"strconv"
)

// Config is a struct that represents the JSON config data for a sk8 client.
type Config struct {
	// Debug instructs the sk8 script to run in debug mode. This
	// causes `set -x` and sets the log level to DEBUG.
	Debug bool `json:"debug,omitempty"`

	// Sk8ScriptPath is the path to the sk8 script. This may be a local
	// file or a URL. If no value is provided then an embedded copy of
	// the script is uploaded to the nodes via cloud-init.
	Sk8ScriptPath string `json:"sk8-script-path,omitempty"`

	// Network is the network configuration.
	Network NetworkConfig `json:"network,omitempty"`

	// Nodes is the configuration for the members of the cluster.
	Nodes NodeConfig `json:"nodes,omitempty"`

	// SSH is the configuration for accessing the nodes remotely via SSH.
	SSH SSHConfig `json:"ssh,omitempty"`

	// VSphere is the information used to access the vSphere endpoint
	// on which the cluster member VMs will be created.
	VSphere VSphereConfig `json:"vsphere,omitempty"`

	// VCenterSimulatorEnabled indicates whether or not to use the
	// vCenter simulator as the endpoint for the vSphere cloud provider.
	// This flag is ignored if the cloud provider is not configured for
	// the in-tree or external vSphere simulator.
	VCenterSimulatorEnabled bool `json:"vcsim,omitempty"`

	// TLS is the TLS configuration.
	TLS TLSConfig `json:"tls,omitempty"`

	// K8s is the kubernetes configuration.
	K8s KubernetesConfig `json:"k8s,omitempty"`

	// Env is a map of environment variable key/value pairs that are written
	// to /etc/default/sk8 and are used to configure the sk8 script.
	Env map[string]string `json:"env,omitempty"`
}

// Build should be invoked before a call to Environ(). This function
// sets default values for missing configuration properties as well as
// validates the configuration data.
//
// This function is not safe for concurrent use and should not be
// called more than once.
func (c *Config) Build(ctx context.Context) error {
	if err := c.readEnv(ctx); err != nil {
		return err
	}
	if err := c.setDefaults(ctx); err != nil {
		return err
	}
	if err := c.validate(ctx); err != nil {
		return err
	}
	return nil
}

// Environ returns the environment varialble map that is used to write
// a defaults file for the sk8 script. The map is generated every time
// the function is called, so be nice.
func (c *Config) Environ(ctx context.Context) (map[string]string, error) {

	env := map[string]string{}
	if c.Debug {
		env["DEBUG"] = "true"
	}

	// Copy the key/value pairs from the config's own
	// environment variable map.
	for k, v := range c.Env {
		env[k] = v
	}

	if err := c.VSphere.setEnv(ctx, env); err != nil {
		return nil, err
	}
	if err := c.Network.setEnv(ctx, env); err != nil {
		return nil, err
	}
	if err := c.Nodes.setEnv(ctx, env); err != nil {
		return nil, err
	}
	if err := c.TLS.setEnv(ctx, env); err != nil {
		return nil, err
	}
	if err := c.K8s.setEnv(ctx, env); err != nil {
		return nil, err
	}
	return env, nil
}

func (c *Config) readEnv(ctx context.Context) error {
	if ok, _ := strconv.ParseBool(os.Getenv("SK8_DEBUG")); ok {
		c.Debug = true
	}
	// All environment variables beginning with SK8S_ENV_ are stripped
	// of that prefix and then inserted into the config's env var map.
	sk8sEnvRX := regexp.MustCompile(`^SK8S_ENV_([^=]+)=(.*)$`)
	for _, v := range os.Environ() {
		if m := sk8sEnvRX.FindStringSubmatch(v); len(m) > 0 {
			if _, ok := c.Env[m[1]]; !ok {
				c.Env[m[1]] = m[2]
			}
		}
	}

	if err := c.VSphere.readEnv(ctx); err != nil {
		return err
	}
	if err := c.SSH.readEnv(ctx); err != nil {
		return err
	}
	if err := c.Network.readEnv(ctx); err != nil {
		return err
	}
	if err := c.Nodes.readEnv(ctx); err != nil {
		return err
	}
	if err := c.TLS.readEnv(ctx); err != nil {
		return err
	}
	if err := c.K8s.readEnv(ctx); err != nil {
		return err
	}

	return nil
}

func (c Config) validate(ctx context.Context) error {
	if err := c.VSphere.validate(ctx); err != nil {
		return err
	}
	if err := c.SSH.validate(ctx); err != nil {
		return err
	}
	if err := c.Network.validate(ctx); err != nil {
		return err
	}
	if err := c.Nodes.validate(ctx); err != nil {
		return err
	}
	if err := c.TLS.validate(ctx); err != nil {
		return err
	}
	if err := c.K8s.validate(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Config) setDefaults(ctx context.Context) error {
	if err := c.VSphere.setDefaults(ctx, *c); err != nil {
		return err
	}
	if err := c.SSH.setDefaults(ctx, *c); err != nil {
		return err
	}
	if err := c.Network.setDefaults(ctx, *c); err != nil {
		return err
	}
	if err := c.Nodes.setDefaults(ctx, *c); err != nil {
		return err
	}
	if err := c.TLS.setDefaults(ctx, *c); err != nil {
		return err
	}
	if err := c.K8s.setDefaults(ctx, *c); err != nil {
		return err
	}
	return nil
}
