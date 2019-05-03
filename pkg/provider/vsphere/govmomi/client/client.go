/*
simple-kubernetes-test-environment

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"

	"vmware.io/sk8/pkg/provider/vsphere/config"
)

// Client is a vSphere client.
type Client struct {
	Client       govmomi.Client
	SOAP         *soap.Client
	Vim25        *vim25.Client
	Rest         *rest.Client
	URL          *url.URL
	Finder       *find.Finder
	Datacenter   *object.Datacenter
	Datastore    *object.Datastore
	Folder       *object.Folder
	ResourcePool *object.ResourcePool
	Networks     map[string]object.NetworkReference
	restURL      string
}

// New returns a new client connection for the provided vSphere endpoint.
func New(
	ctx context.Context,
	cfg config.ClusterProviderConfig) (*Client, error) {

	serverURL, err := soap.ParseURL(cfg.Server)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error parsing vSphere server: %s", cfg.Server)
	}
	serverURL.User = nil
	log.WithField("url", serverURL).Debug("parsed vSphere server URL")

	soapClient := soap.NewClient(serverURL, false)
	soapClient.Namespace = "urn:" + vim25.Namespace
	soapClient.Version = "5.5"
	soapClient.UserAgent = "sk8"
	roundTripper := vim25.Retry(soapClient, vim25.TemporaryNetworkError(3))

	vim25Client, err := vim25.NewClient(ctx, roundTripper)
	if err != nil {
		return nil, errors.Wrap(err, "error creating vim25 client")
	}
	vim25Client.Client = soapClient

	sessionManager := session.NewManager(vim25Client)
	userInfo := url.UserPassword(cfg.Username, cfg.Password)
	if err := sessionManager.Login(ctx, userInfo); err != nil {
		return nil, errors.Wrap(err, "error logging into vSphere")
	}

	restClient := rest.NewClient(vim25Client)
	if err := restClient.Login(ctx, userInfo); err != nil {
		return nil, errors.Wrap(err, "error logging into vSphere REST API")
	}

	return &Client{
		Client: govmomi.Client{
			Client:         vim25Client,
			SessionManager: sessionManager,
		},
		SOAP:    soapClient,
		Vim25:   vim25Client,
		Rest:    restClient,
		URL:     serverURL,
		restURL: strings.Replace(serverURL.String(), "/sdk", "", 1),
	}, nil
}

// WithMachineProviderConfig uses the given cfg to assign the client's
// datacenter, datastore, etc. fields.
func (c *Client) WithMachineProviderConfig(
	ctx context.Context,
	cfg config.MachineProviderConfig) (*Client, error) {

	finder := find.NewFinder(c.Vim25, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, cfg.Datacenter)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error finding datacenter %s", cfg.Datacenter)
	}
	finder.SetDatacenter(datacenter)
	datastore, err := finder.Datastore(ctx, cfg.Datastore)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error finding datastore %s", cfg.Datastore)
	}
	folder, err := finder.Folder(ctx, cfg.Folder)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error finding folder %s", cfg.Folder)
	}
	pool, err := finder.ResourcePool(ctx, cfg.ResourcePool)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error finding resource pool %s", cfg.ResourcePool)
	}

	// Validate that the provided network names exist.
	networks := map[string]object.NetworkReference{}
	for _, iface := range cfg.Network.Interfaces {
		name := iface.Network
		ref, err := finder.Network(ctx, name)
		if err != nil {
			return nil, errors.Wrapf(err, "error finding network %q", name)
		}
		networks[name] = ref
	}

	c.Finder = finder
	c.Datacenter = datacenter
	c.Datastore = datastore
	c.Folder = folder
	c.Networks = networks
	c.ResourcePool = pool

	return c, nil
}

// GetRestURL gets the URL for a REST resource.
func (c *Client) GetRestURL(suffix string) string {
	if strings.HasPrefix(suffix, "/") {
		return fmt.Sprintf("%s%s", c.restURL, suffix)
	}
	return fmt.Sprintf("%s/%s", c.restURL, suffix)
}
