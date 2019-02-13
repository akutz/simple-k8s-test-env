package app // import "vmw.io/sk8/app"

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	defaultTemplateName          = "photon-cloud-init"
	defaultLibraryName           = "sk8"
	defaultLibraryItemName       = "photon2-cloud-init"
	defaultPhotonCloudInitOVAURL = "https://s3-us-west-2.amazonaws.com/cnx.vmware/photon2-cloud-init.ova"
)

// VSphereConfig is the information used to access a vSphere endpoint.
type VSphereConfig struct {
	// ServerURL is used to specify the complete URL to the SDK endpoint.
	// This value should adhere to the following format:
	//
	//     SCHEME://SERVER[:PORT]/sdk
	//
	// The SCHEME value may be set to either "http" or "https", SERVER
	// should be the IP address or FQDN of a vSphere server, and a PORT
	// may also be specified. Please note that the URL should end with
	// "/sdk", otherwise requests will fail.
	ServerURL string `json:"url,omitempty"`
	serverURL url.URL

	// Username and Password are the credentials used to access the vSphere
	// server. Please note that these should not be included as part of the
	// URL.
	Username string `json:"-"`
	Password string `json:"-"`

	// Insecure is a flag that indicates whether or not to verify peer
	// certificates for TLS connections.
	Insecure bool `json:"insecure,omitempty"`

	// PhotonCloudInitOVAURL is the path to a Photon OVA with cloud-init and the
	// the VMware GuestInfo datasource enabled. The default value is
	// https://s3-us-west-2.amazonaws.com/cnx.vmware/photon2-cloud-init.ova.
	PhotonCloudInitOVAURL string `json:"photon-cloud-init-ova-url,omitempty"`
	photonCloudInitOVAURL url.URL

	// Template is the path to the Photon2 template used to create new VMs.
	// If a template is unavailable then the content library is used. Please
	// note that only vSphere 6.0+ supports content library.
	Template string `json:"template,omitempty"`

	// TemplateDisabled may be set to true in order to bypass the template
	// altogether in favor of using the content library.
	TemplateDisabled bool `json:"template-disabled"`

	// LibraryName and LibraryItem are used when a Template is not availble
	// for cloning. If the provided library or item are unavailable then
	// they are created. The Photon OVA specified by PhotonCloudInitURL is
	// then downloaded into the the library item as an OVF used to deploy VMs.
	LibraryName     string `json:"lib-name,omitempty"`
	LibraryItemName string `json:"lib-item-name,omitempty"`

	// Information about where the node VMs will be created and to what
	// network they will be attached.
	Datacenter   string `json:"datacenter,omitempty"`
	Folder       string `json:"folder,omitempty"`
	Datastore    string `json:"datastore,omitempty"`
	ResourcePool string `json:"resource-pool,omitempty"`
	Network      string `json:"network,omitempty"`
}

func (c *VSphereConfig) port() string {
	p := c.serverURL.Port()
	if p != "" {
		return p
	}
	if strings.EqualFold("https", c.serverURL.Scheme) {
		return "443"
	}
	return "80"
}

func (c *VSphereConfig) readEnv(ctx context.Context) error {
	if c.Template == "" {
		c.Template = os.Getenv("SK8_VSPHERE_TEMPLATE")
	}
	c.TemplateDisabled, _ = strconv.ParseBool(
		os.Getenv("SK8_VSPHERE_TEMPLATE_DISABLED"))

	if c.LibraryName == "" {
		c.LibraryName = os.Getenv("SK8_VSPHERE_LIBRARY_NAME")
	}
	if c.LibraryItemName == "" {
		c.LibraryItemName = os.Getenv("SK8_VSPHERE_LIBRARY_ITEM_NAME")
	}
	if c.PhotonCloudInitOVAURL == "" {
		c.PhotonCloudInitOVAURL = os.Getenv("SK8_VSPHERE_PHOTON_CLOUD_INIT_OVA_URL")
	}
	if c.Datacenter == "" {
		c.Datacenter = os.Getenv("SK8_VSPHERE_DATACENTER")
	}
	if c.Datastore == "" {
		c.Datastore = os.Getenv("SK8_VSPHERE_DATASTORE")
	}
	if c.Folder == "" {
		c.Folder = os.Getenv("SK8_VSPHERE_FOLDER")
	}
	c.Insecure, _ = strconv.ParseBool(os.Getenv("SK8_VSPHERE_INSECURE"))
	if c.Network == "" {
		c.Network = os.Getenv("SK8_VSPHERE_NETWORK")
	}
	if c.ResourcePool == "" {
		c.ResourcePool = os.Getenv("SK8_VSPHERE_RESOURCE_POOL")
	}
	if c.ServerURL == "" {
		c.ServerURL = os.Getenv("SK8_VSPHERE_URL")
	}
	if c.Username == "" {
		if c.Username = os.Getenv("SK8_VSPHERE_USER"); c.Username == "" {
			c.Username = os.Getenv("SK8_VSPHERE_USERNAME")
		}
	}
	if c.Password == "" {
		c.Password = os.Getenv("SK8_VSPHERE_PASSWORD")
	}

	return nil
}

func (c *VSphereConfig) validate(ctx context.Context) error {
	if c.Username == "" {
		return fmt.Errorf("vsphere.username required")
	}
	if c.Password == "" {
		return fmt.Errorf("vsphere.password required")
	}
	if c.ServerURL == "" {
		return fmt.Errorf("vsphere.server-url required")
	}
	return nil
}

func (c *VSphereConfig) setDefaults(ctx context.Context, cfg Config) error {
	{
		u, err := url.Parse(c.ServerURL)
		if err != nil {
			return err
		}
		if u.Path != "/sdk" {
			u.Path = "/sdk"
		}
		c.serverURL = *u
		c.ServerURL = u.String()
	}
	{
		if c.PhotonCloudInitOVAURL == "" {
			c.PhotonCloudInitOVAURL = defaultPhotonCloudInitOVAURL
		}
		u, err := url.Parse(c.PhotonCloudInitOVAURL)
		if err != nil {
			return err
		}
		c.photonCloudInitOVAURL = *u
		c.PhotonCloudInitOVAURL = u.String()
	}
	if c.Template == "" {
		c.Template = defaultTemplateName
	}
	if c.LibraryName == "" {
		c.LibraryName = defaultLibraryName
	}
	if c.LibraryItemName == "" {
		c.LibraryItemName = defaultLibraryItemName
	}
	return nil
}

func (c *VSphereConfig) setEnv(
	ctx context.Context, env map[string]string) error {
	return nil
}
