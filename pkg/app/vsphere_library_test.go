package app_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/vmware/govmomi/vapi/vcenter"

	"github.com/vmware/govmomi/object"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vapi/library"

	"github.com/vmware/govmomi/vapi/rest"

	"github.com/vmware/sk8/pkg/app"
	"github.com/vmware/sk8/pkg/config"
)

const ovaURL = "https://s3-us-west-2.amazonaws.com/cnx.vmware/cicd/photon2-cloud-init.ova"

func TestUploadOVA(t *testing.T) {
	cfg := config.Config{}
	ctx := context.Background()
	if err := app.ValidateConfig(ctx, &cfg); err != nil {
		t.Fatal(err)
	}

	// Get a client for the vSphere endpoint.
	client, err := app.NewVSphereClient(ctx, cfg.VSphere)
	if err != nil {
		t.Fatal(err)
	}

	// Get the REST client.
	restClient := rest.NewClient(client.Client)
	userInfo := url.UserPassword(cfg.VSphere.Username, cfg.VSphere.Password)
	if err := restClient.Login(ctx, userInfo); err != nil {
		t.Fatal(err)
	}

	// Get the content library client.
	libClient := app.NewLibraryClient(restClient, os.Stdout)

	// Upload the OVA
	item, err := uploadOVA(ctx, libClient)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("item name=%s id=%s", item.Name, item.ID)
}

func TestDeployOVF(t *testing.T) {
	cfg := config.Config{}
	ctx := context.Background()
	if err := app.ValidateConfig(ctx, &cfg); err != nil {
		t.Fatal(err)
	}

	// Get a client for the vSphere endpoint.
	client, err := app.NewVSphereClient(ctx, cfg.VSphere)
	if err != nil {
		t.Fatal(err)
	}

	// Get the REST client.
	restClient := rest.NewClient(client.Client)
	userInfo := url.UserPassword(cfg.VSphere.Username, cfg.VSphere.Password)
	if err := restClient.Login(ctx, userInfo); err != nil {
		t.Fatal(err)
	}

	// Get the content library client.
	libClient := app.NewLibraryClient(restClient, os.Stdout)

	// Upload the OVA
	item, err := uploadOVA(ctx, libClient)
	if err != nil {
		t.Fatal(err)
	}

	// Find the objects needed to create the VMs.
	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(
		ctx, cfg.VSphere.Datacenter)
	if err != nil {
		t.Fatal(err)
	}
	finder.SetDatacenter(datacenter)
	datastore, err := finder.Datastore(ctx, cfg.VSphere.Datastore)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := finder.ResourcePool(ctx, cfg.VSphere.ResourcePool)
	if err != nil {
		t.Fatal(err)
	}
	folder, err := finder.Folder(ctx, cfg.VSphere.Folder)
	if err != nil {
		t.Fatal(err)
	}

	// Deploy the OVF
	vmRef, err := libClient.DeployOVF(
		ctx, item.ID,
		vcenter.Deploy{
			Target: vcenter.Target{
				ResourcePoolID: pool.Reference().Value,
				FolderID:       folder.Reference().Value,
			},
			DeploymentSpec: vcenter.DeploymentSpec{
				Name:               "sk8-TestDeployOVF",
				AcceptAllEULA:      true,
				DefaultDatastoreID: datastore.Reference().Value,
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	vm := object.NewVirtualMachine(client.Client, *vmRef)
	vmName, err := vm.ObjectName(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("vm name=%s uuid=%s", vmName, vm.UUID(ctx))
}

func uploadOVA(
	ctx context.Context,
	client *app.LibraryClient) (*library.Item, error) {

	return client.UploadOVA(
		ctx,
		library.Library{
			Name: "sk8-TestUploadOVA",
			Storage: []library.StorageBackings{
				library.StorageBackings{
					Type:        "DATASTORE",
					DatastoreID: "datastore-61",
				},
			},
		},
		library.Item{
			Name: "photon2-cloud-init",
			Type: "ovf",
		},
		"https://s3-us-west-2.amazonaws.com/cnx.vmware/photon2-cloud-init.ova",
		true)
}
