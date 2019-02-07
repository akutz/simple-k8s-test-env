package app // import "github.com/vmware/sk8/pkg/app"

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware/sk8/pkg/config"
)

// NewVSphereClient returns a new client connection for the provided
// vSphere endpoint configuration.
func NewVSphereClient(
	ctx context.Context,
	c config.VSphereConfig) (*govmomi.Client, error) {

	serverURL, err := getServerURL(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("parse url failed: %v", err)
	}
	serverURL.User = nil

	soapClient := soap.NewClient(serverURL, c.Insecure)
	roundTripper, err := configureRoundTripper(ctx, soapClient)
	if err != nil {
		return nil, fmt.Errorf("round tripper failed: %v", err)
	}

	client, err := vim25.NewClient(ctx, roundTripper)
	if err != nil {
		return nil, fmt.Errorf("new client failed: %v", err)
	}
	client.Client = soapClient

	sessionManager := session.NewManager(client)
	userInfo := url.UserPassword(c.Username, c.Password)
	if err := sessionManager.Login(ctx, userInfo); err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}

	clientHelper := &govmomi.Client{
		Client:         client,
		SessionManager: sessionManager,
	}

	return clientHelper, nil
}

func configureRoundTripper(
	ctx context.Context, sc *soap.Client) (soap.RoundTripper, error) {

	// Set namespace and version
	sc.Namespace = "urn:" + vim25.Namespace
	sc.Version = "5.5"
	sc.UserAgent = "sk8"

	// Retry twice when a temporary I/O error occurs.
	// This means a maximum of 3 attempts.
	return vim25.Retry(sc, vim25.TemporaryNetworkError(3)), nil
}

// GetRESTURL gets a REST URL.
func GetRESTURL(
	client *rest.Client,
	suffix string) string {

	return fmt.Sprintf("%s%s",
		strings.Replace(client.URL().String(), "/sdk", "", 1),
		suffix)
}

func getDiskDeviceConfigSpec(
	ctx context.Context,
	sizeGiB uint32,
	devs object.VirtualDeviceList) (types.BaseVirtualDeviceConfigSpec, error) {

	// Search for the first disk and update its size.
	for _, dev := range devs {
		if disk, ok := dev.(*types.VirtualDisk); ok {
			disk.CapacityInKB = int64(sizeGiB) * 1024 * 1024
			return &types.VirtualDeviceConfigSpec{
				Operation: types.VirtualDeviceConfigSpecOperationEdit,
				Device:    disk,
			}, nil
		}
	}

	return nil, fmt.Errorf("no disk found")
}

func getNetworkDeviceConfigSpec(
	ctx context.Context,
	netw object.NetworkReference,
	devs object.VirtualDeviceList) (types.BaseVirtualDeviceConfigSpec, error) {

	// Prepare virtual device config spec for network card.
	op := types.VirtualDeviceConfigSpecOperationAdd
	card, err := networkDevice(ctx, netw)
	if err != nil {
		return nil, err
	}

	// Search for the first network card of the source and update
	// the config spec's network device if necessary.
	for _, dev := range devs {
		if _, ok := dev.(types.BaseVirtualEthernetCard); ok {
			op = types.VirtualDeviceConfigSpecOperationEdit
			changeNetDevice(dev, card)
			card = dev
			break
		}
	}
	return &types.VirtualDeviceConfigSpec{
		Operation: op,
		Device:    card,
	}, nil
}

func networkDevice(
	ctx context.Context,
	network object.NetworkReference) (types.BaseVirtualDevice, error) {

	backing, err := network.EthernetCardBackingInfo(ctx)
	if err != nil {
		return nil, err
	}
	dev, err := object.EthernetCardTypes().CreateEthernetCard("e1000", backing)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

func changeNetDevice(from types.BaseVirtualDevice, to types.BaseVirtualDevice) {
	current := from.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
	changed := to.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
	current.Backing = changed.Backing
	if changed.MacAddress != "" {
		current.MacAddress = changed.MacAddress
	}
	if changed.AddressType != "" {
		current.AddressType = changed.AddressType
	}
}

func readVSphereConfigEnv(ctx context.Context, c *config.VSphereConfig) {
	if c.Template == "" {
		c.Template = os.Getenv("VSPHERE_TEMPLATE")
		if c.Template == "" {
			c.Template = "photon2-cloud-init"
		}
	}
	if c.Datacenter == "" {
		c.Datacenter = os.Getenv("VSPHERE_DATACENTER")
	}
	if c.Datastore == "" {
		c.Datastore = os.Getenv("VSPHERE_DATASTORE")
	}
	if c.Folder == "" {
		c.Folder = os.Getenv("VSPHERE_FOLDER")
	}
	if !c.Insecure {
		c.Insecure, _ = strconv.ParseBool(os.Getenv("VSPHERE_INSECURE"))
	}
	if c.Network == "" {
		c.Network = os.Getenv("VSPHERE_NETWORK")
	}
	if c.Password == "" {
		c.Password = os.Getenv("VSPHERE_PASSWORD")
	}
	if c.Port == 0 {
		i, _ := strconv.Atoi(os.Getenv("VSPHERE_SERVER_PORT"))
		c.Port = uint32(i)
	}
	if c.ResourcePool == "" {
		c.ResourcePool = os.Getenv("VSPHERE_RESOURCE_POOL")
	}
	if c.Server == "" {
		c.Server = os.Getenv("VSPHERE_SERVER")
	}
	if c.Username == "" {
		if c.Username = os.Getenv("VSPHERE_USER"); c.Username == "" {
			c.Username = os.Getenv("VSPHERE_USERNAME")
		}
	}
}

func validateVSphereConfig(ctx context.Context, c *config.VSphereConfig) error {
	if c.Template == "" {
		return fmt.Errorf("vsphere.template required")
	}
	if c.Datacenter == "" {
		return fmt.Errorf("vsphere.datacenter required")
	}
	if c.Datastore == "" {
		return fmt.Errorf("vsphere.datastore required")
	}
	if c.Folder == "" {
		return fmt.Errorf("vsphere.folder required")
	}
	if c.Network == "" {
		return fmt.Errorf("vsphere.network required")
	}
	if c.Password == "" {
		return fmt.Errorf("vsphere.password required")
	}
	if c.ResourcePool == "" {
		return fmt.Errorf("vsphere.resource-pool required")
	}
	if c.Server == "" {
		return fmt.Errorf("vsphere.server required")
	}
	if c.Username == "" {
		return fmt.Errorf("vsphere.username required")
	}
	return nil
}

func GetServerURL(
	ctx context.Context, c config.VSphereConfig) (*url.URL, error) {
	return getServerURL(ctx, c)
}

func getServerURL(
	ctx context.Context, c config.VSphereConfig) (*url.URL, error) {

	scheme := "http"
	if strings.HasPrefix(c.Server, "http://") {
		scheme = ""
		if c.Port == 0 {
			c.Port = 80
		}
	} else if strings.HasPrefix(c.Server, "https://") {
		scheme = ""
		if c.Port == 0 {
			c.Port = 443
		}
	} else if c.Port == 80 {
		scheme = "http://"
	} else {
		scheme = "https://"
		if c.Port == 0 {
			c.Port = 443
		}
	}

	c.Server = fmt.Sprintf("%s%s:%d", scheme, c.Server, c.Port)
	if !(strings.HasSuffix(c.Server, "/sdk") ||
		strings.HasSuffix(c.Server, "/sdk/")) {
		c.Server = fmt.Sprintf("%s/sdk", c.Server)
	}

	return url.Parse(c.Server)
}

func getServerHostAndPort(
	ctx context.Context, c config.VSphereConfig) (string, string, error) {

	serverURL, err := getServerURL(ctx, c)
	if err != nil {
		return "", "", err
	}
	return serverURL.Hostname(), serverURL.Port(), nil
}

// ExtraConfig is data used with a VM's guestInfo RPC interface.
type ExtraConfig []types.BaseOptionValue

// SetCloudInitUserData sets the cloud init user data at the key
// "guestinfo.userdata" as a gzipped, base64-encoded string.
func (e *ExtraConfig) SetCloudInitUserData(data []byte) error {

	encData, err := Base64GzipBytes(data)
	if err != nil {
		return err
	}

	*e = append(*e,
		&types.OptionValue{
			Key:   "guestinfo.userdata",
			Value: encData,
		},
		&types.OptionValue{
			Key:   "guestinfo.userdata.encoding",
			Value: "gzip+base64",
		},
	)

	return nil
}

// SetCloudInitMetadata sets the cloud init user data at the key
// "guestinfo.metadata" as a gzipped, base64-encoded string.
func (e *ExtraConfig) SetCloudInitMetadata(data []byte) error {

	encData, err := Base64GzipBytes(data)
	if err != nil {
		return err
	}

	*e = append(*e,
		&types.OptionValue{
			Key:   "guestinfo.metadata",
			Value: encData,
		},
		&types.OptionValue{
			Key:   "guestinfo.metadata.encoding",
			Value: "gzip+base64",
		},
	)

	return nil
}

func initCloudProviderConfig(ctx context.Context, c *config.Config) error {
	if c.VCenterSimulatorEnabled {
		c.Env["VCSIM"] = "true"
	}
	switch c.K8s.CloudProvider.Type {
	case config.InTreeCloudProvider:
		if c.VCenterSimulatorEnabled {
			c.Env["CLOUD_PROVIDER"] = "vsphere"
			return nil
		}
		return initInTreeCloudProviderConfig(ctx, c)
	case config.ExternalCloudProvider:
		if c.VCenterSimulatorEnabled {
			c.Env["CLOUD_PROVIDER"] = "external"
			c.Env["CLOUD_PROVIDER_EXTERNAL"] = "vsphere"
			return nil
		}
		return initExternalCloudProviderConfig(ctx, c)
	default:
		return nil
	}
}
func initInTreeCloudProviderConfig(
	ctx context.Context, c *config.Config) error {

	host, port, err := getServerHostAndPort(ctx, c.VSphere)
	if err != nil {
		return err
	}

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
		c.VSphere.Username,
		c.VSphere.Password,
		port,
		c.VSphere.Insecure,
		c.VSphere.Datacenter,
		host,
		c.VSphere.Folder,
		c.VSphere.Datastore,
		c.VSphere.ResourcePool,
		c.VSphere.Network,
	}); err != nil {
		return err
	}

	bufEnc, err := Base64GzipBytes(buf.Bytes())
	if err != nil {
		return err
	}

	c.Env["CLOUD_PROVIDER"] = "vsphere"
	c.Env["CLOUD_CONFIG"] = bufEnc

	return nil
}
func initExternalCloudProviderConfig(
	ctx context.Context, c *config.Config) error {

	host, port, err := getServerHostAndPort(ctx, c.VSphere)
	if err != nil {
		return err
	}

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
			port,
			c.VSphere.Insecure,
			c.VSphere.Datacenter,
			host,
		}); err != nil {
			return err
		}

		bufEnc, err := Base64GzipBytes(buf.Bytes())
		if err != nil {
			return err
		}

		c.Env["CLOUD_CONFIG"] = bufEnc
	}

	{
		tpl := template.Must(
			template.New("t").Parse(externalCloudProviderSecretsFormat))

		usernameBase64 := base64.StdEncoding.EncodeToString(
			[]byte(c.VSphere.Username))
		passwordBase64 := base64.StdEncoding.EncodeToString(
			[]byte(c.VSphere.Password))
		buf := &bytes.Buffer{}
		if err := tpl.Execute(buf, struct {
			Server         string
			UsernameBase64 string
			PasswordBase64 string
		}{
			host,
			usernameBase64,
			passwordBase64,
		}); err != nil {
			return err
		}

		bufEnc, err := Base64GzipBytes(buf.Bytes())
		if err != nil {
			return err
		}

		c.Env["MANIFEST_YAML_AFTER_RBAC_2"] = bufEnc
	}

	c.Env["CLOUD_PROVIDER"] = "external"
	c.Env["CLOUD_PROVIDER_EXTERNAL"] = "vsphere"

	return nil
}

// DownloadOVA downloads an OVA to a local file.
func DownloadOVA(
	ctx context.Context,
	out io.Writer,
	c config.VSphereConfig) (string, error) {

	if out == nil {
		out = ioutil.Discard
	}

	ovaFilePath := c.TemplateOVA
	if strings.HasPrefix("http", c.TemplateOVA) {
		ovaFilePath = "photon2-cloud-init.ova"
		if FileExists(ovaFilePath) {
			fmt.Fprintf(out, "using local cache %s...\n", ovaFilePath)
			return ovaFilePath, nil
		}

		resp, err := http.Get(c.TemplateOVA)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(out, "downloading ova %s...", c.TemplateOVA)
		defer resp.Body.Close()
		f, err := os.Create("photon2-cloud-init.ova")
		defer f.Close()
		if _, err := io.Copy(
			f, io.TeeReader(resp.Body, &writeCounter{w: out})); err != nil {
			fmt.Fprintf(out, "failed!\n")
			return "", err
		}
		fmt.Fprintf(out, "success!\n")
	}

	return ovaFilePath, nil
}

type writeCounter struct {
	w io.Writer
}

func (w *writeCounter) Write(p []byte) (int, error) {
	fmt.Fprint(w.w, ".")
	return len(p), nil
}

func uploadMissingCloneSource(
	ctx context.Context,
	out io.Writer,
	c *config.Config,
	client *govmomi.Client) (*object.VirtualMachine, error) {

	// Download the OVA if the provided path is remote.
	//ovaFilePath, err := DownloadOVA(ctx, out, c.VSphere)
	//if err != nil {
	//	return nil, err
	//}

	//ovfm := ovf.NewManager(client)
	ovaURL := c.VSphere.TemplateOVA
	if len(ovaURL) == 0 {
		ovaURL = `https://s3-us-west-2.amazonaws.com/cnx.vmware/cicd/photon2-cloud-init.ova`
	}
	u, err := url.Parse(ovaURL)
	if err != nil {
		return nil, err
	}
	_ = u
	//r, n, err := client.Download(ctx, u, &soap.DefaultDownload)

	return nil, nil
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
