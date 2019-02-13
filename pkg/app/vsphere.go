package app // import "vmw.io/sk8/app"

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// NewVSphereClient returns a new client connection for the provided
// vSphere endpoint configuration.
func NewVSphereClient(
	ctx context.Context,
	c VSphereConfig) (*govmomi.Client, error) {

	soapClient := soap.NewClient(&c.serverURL, c.Insecure)
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

