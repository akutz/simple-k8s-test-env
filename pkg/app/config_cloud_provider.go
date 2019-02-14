package app // import "vmw.io/sk8/app"

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"strconv"
	"text/template"
)

const (
	vsphereCloudProvider      = "vsphere"
	externalCloudProvider     = "external"
	defaultCloudProviderImage = "gcr.io/cloud-provider-vsphere/vsphere-cloud-controller-manager:v0.1.0"
)

// CloudProviderConfig is the data used to configure the vSphere cloud provider.
type CloudProviderConfig struct {
	// An in-tree or external cloud provider may be configured. The
	// CloudProviderConfig is required for an in-tree cloud provider but
	// may be omitted in place of the Manifests field when configuring
	// an external cloud provider. The log level applies only to
	// external cloud providers with valid values 1-10. The default
	// log level is 2.
	Type     CloudProviderType `json:"type,omitempty"`
	Image    string            `json:"image,omitempty"`
	Config   []byte            `json:"config,omitempty"`
	LogLevel uint8             `json:"log-level,omitempty"`

	config        string
	configSecrets string
}

// CloudProviderType is the type for the enumeration that defines the
// possible choices for cloud provider.
type CloudProviderType uint8

const (
	// UnknownCloudProvider indicates no choice has been made.
	UnknownCloudProvider CloudProviderType = iota

	// NoCloudProvider disables the use of a cloud provider entirely.
	NoCloudProvider

	// InTreeCloudProvider specifies the use of an in-tree cloud provider.
	InTreeCloudProvider

	// ExternalCloudProvider indicates the use of an external cloud provider.
	ExternalCloudProvider
)

// ParseCloudProviderType returns a cloud provider type value.
func ParseCloudProviderType(s string) CloudProviderType {
	i, _ := strconv.ParseUint(s, 10, 8)
	switch CloudProviderType(uint8(i)) {
	case NoCloudProvider:
		return NoCloudProvider
	case InTreeCloudProvider:
		return InTreeCloudProvider
	case ExternalCloudProvider:
		return ExternalCloudProvider
	default:
		return UnknownCloudProvider
	}
}

func (c *CloudProviderConfig) readEnv(ctx context.Context) error {
	if c.Type == UnknownCloudProvider {
		c.Type = ParseCloudProviderType(os.Getenv("SK8_CLOUD_PROVIDER_TYPE"))
	}
	if c.Image == "" {
		c.Image = os.Getenv("SK8_CLOUD_PROVIDER_IMAGE")
	}
	return nil
}

func (c *CloudProviderConfig) validate(ctx context.Context) error {
	return nil
}

func (c *CloudProviderConfig) setDefaults(
	ctx context.Context, cfg Config) error {

	if c.Type == UnknownCloudProvider {
		c.Type = ExternalCloudProvider
	}
	if c.Type == ExternalCloudProvider && c.Image == "" {
		c.Image = defaultCloudProviderImage
	}

	if !cfg.VCenterSimulatorEnabled {
		switch c.Type {
		case InTreeCloudProvider:
			if err := c.initInTree(ctx, cfg); err != nil {
				return err
			}
		case ExternalCloudProvider:
			if err := c.initExternal(ctx, cfg); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *CloudProviderConfig) initInTree(
	ctx context.Context, cfg Config) error {

	tpl := template.Must(
		template.New("t").Parse(intreeCloudProviderConfigFormat))
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, struct {
		Username     string
		Password     string
		Port         string
		Insecure     bool
		Datacenter   string
		Server       string
		Folder       string
		Datastore    string
		ResourcePool string
		Network      string
	}{
		cfg.VSphere.Username,
		cfg.VSphere.Password,
		cfg.VSphere.port(),
		cfg.VSphere.Insecure,
		cfg.VSphere.Datacenter,
		cfg.VSphere.serverURL.Hostname(),
		cfg.VSphere.Folder,
		cfg.VSphere.Datastore,
		cfg.VSphere.ResourcePool,
		cfg.VSphere.Network,
	}); err != nil {
		return err
	}

	bufEnc, err := Base64GzipBytes(buf.Bytes())
	if err != nil {
		return err
	}
	c.config = bufEnc
	return nil
}

func (c *CloudProviderConfig) initExternal(
	ctx context.Context, cfg Config) error {

	{
		tpl := template.Must(
			template.New("t").Parse(externalCloudProviderConfigFormat))

		buf := &bytes.Buffer{}
		if err := tpl.Execute(buf, struct {
			Port       string
			Insecure   bool
			Datacenter string
			Server     string
		}{
			cfg.VSphere.port(),
			cfg.VSphere.Insecure,
			cfg.VSphere.Datacenter,
			cfg.VSphere.serverURL.Hostname(),
		}); err != nil {
			return err
		}

		bufEnc, err := Base64GzipBytes(buf.Bytes())
		if err != nil {
			return err
		}
		c.config = bufEnc
	}

	{
		tpl := template.Must(
			template.New("t").Parse(externalCloudProviderSecretsFormat))

		usernameBase64 := base64.StdEncoding.EncodeToString(
			[]byte(cfg.VSphere.Username))
		passwordBase64 := base64.StdEncoding.EncodeToString(
			[]byte(cfg.VSphere.Password))
		buf := &bytes.Buffer{}
		if err := tpl.Execute(buf, struct {
			Server         string
			UsernameBase64 string
			PasswordBase64 string
		}{
			cfg.VSphere.serverURL.Hostname(),
			usernameBase64,
			passwordBase64,
		}); err != nil {
			return err
		}

		bufEnc, err := Base64GzipBytes(buf.Bytes())
		if err != nil {
			return err
		}
		c.configSecrets = bufEnc
	}

	return nil
}

func (c *CloudProviderConfig) setEnv(
	ctx context.Context, env map[string]string) error {

	switch c.Type {
	case InTreeCloudProvider:
		env["CLOUD_PROVIDER"] = vsphereCloudProvider
	case ExternalCloudProvider:
		env["CLOUD_PROVIDER"] = externalCloudProvider
		env["CLOUD_PROVIDER_EXTERNAL"] = vsphereCloudProvider
		env["CLOUD_PROVIDER_IMAGE"] = c.Image
	}
	if c.config != "" {
		env["CLOUD_CONFIG"] = c.config
	}
	if c.configSecrets != "" {
		env["MANIFEST_YAML_AFTER_RBAC_2"] = c.configSecrets
	}
	return nil
}

const intreeCloudProviderConfigFormat = `[Global]
user               = "{{.Username}}"
password           = "{{.Passsword}}"
port               = "{{.Port}}"
insecure-flag      = "{{.Insecure}}"
datacenters        = "{{.Datacenter}}"

[VirtualCenter "{{.Server}}"]

[Workspace]
server             = "{{.Server}}"
datacenter         = "{{.Datacenter}}"
folder             = "{{.Folder}}"
default-datastore  = "{{.Datastore}}"
resourcepool-path  = "{{.ResourcePool}}"

[Disk]
scsicontrollertype = pvscsi

[Network]
public-network     = "{{.Network}}"
`

const externalCloudProviderConfigFormat = `[Global]
secret-name        = "cloud-provider-vsphere-credentials"
secret-namespace   = "kube-system"
service-account    = "cloud-controller-manager"
port               = "{{.Port}}"
insecure-flag      = "{{.Insecure}}"
datacenters        = "{{.Datacenter}}"

[VirtualCenter "{{.Server}}"]
`

const externalCloudProviderSecretsFormat = `apiVersion: v1
kind: Secret
metadata:
  name: cloud-provider-vsphere-credentials
  namespace: kube-system
data:
  {{.Server}}.username: "{{.UsernameBase64}}"
  {{.Server}}.password: "{{.PasswordBase64}}"
`
