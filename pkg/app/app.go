package app // import "vmw.io/sk8/app"

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/vcenter"
	"github.com/vmware/govmomi/vim25/types"
)

// Up deploys a Kubernetes cluster using the provided configuration and
// returns the result of the operation. The io.Writer is used to record
// progress. Use ioutil.Discard or nil to ignore progress.
func Up(
	ctx context.Context,
	out io.Writer,
	cfg Config) (*Config, error) {

	if out == nil {
		out = ioutil.Discard
	}

	// Get a client for the vSphere endpoint.
	client, err := NewVSphereClient(ctx, cfg.VSphere)
	if err != nil {
		return nil, err
	}

	// Find the objects needed to create the VMs.
	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(
		ctx, cfg.VSphere.Datacenter)
	if err != nil {
		return nil, err
	}
	finder.SetDatacenter(datacenter)
	datastore, err := finder.Datastore(ctx, cfg.VSphere.Datastore)
	if err != nil {
		return nil, err
	}
	pool, err := finder.ResourcePool(ctx, cfg.VSphere.ResourcePool)
	if err != nil {
		return nil, err
	}
	folder, err := finder.Folder(ctx, cfg.VSphere.Folder)
	if err != nil {
		return nil, err
	}
	network, err := finder.Network(ctx, cfg.VSphere.Network)
	if err != nil {
		return nil, err
	}

	// Get the REST client.
	restClient := rest.NewClient(client.Client)
	userInfo := url.UserPassword(cfg.VSphere.Username, cfg.VSphere.Password)
	if err := restClient.Login(ctx, userInfo); err != nil {
		return nil, err
	}

	// Get the content library client.
	libClient := NewLibraryClient(restClient, os.Stdout)

	// Upload the OVA
	libItem, err := libClient.UploadOVA(
		ctx,
		library.Library{
			Name: cfg.VSphere.LibraryName,
			Storage: []library.StorageBackings{
				library.StorageBackings{
					Type:        "DATASTORE",
					DatastoreID: datastore.Reference().Value,
				},
			},
			Type: "LOCAL",
		},
		library.Item{
			Name: cfg.VSphere.LibraryItemName,
			Type: "ovf",
		},
		cfg.VSphere.PhotonCloudInitOVAURL,
		false)
	if err != nil {
		return nil, err
	}

	var (
		poolID      = pool.Reference().Value
		folderID    = folder.Reference().Value
		datastoreID = datastore.Reference().Value
	)

	var (
		deployWait sync.WaitGroup
		deployDone = make(chan struct{})
		deployErrs = make(chan error)
	)

	deployNode := func(n *SingleNodeConfig) {
		deploySpec := vcenter.Deploy{
			Target: vcenter.Target{
				ResourcePoolID: poolID,
				FolderID:       folderID,
			},
			DeploymentSpec: vcenter.DeploymentSpec{
				Name:               n.Name,
				AcceptAllEULA:      true,
				DefaultDatastoreID: datastoreID,
			},
		}

		// Deploy the VM.
		vmRef, err := libClient.DeployOVF(ctx, libItem.ID, deploySpec)
		if err != nil {
			deployErrs <- err
			return
		}

		n.vm = object.NewVirtualMachine(client.Client, *vmRef)

		// Get the devices for the newly deployed VM.
		devices, err := n.vm.Device(ctx)
		if err != nil {
			deployErrs <- err
			return
		}

		// Search for the first network card of the source and update
		// the config spec's network device if necessary.
		netConfigSpec, err := getNetworkDeviceConfigSpec(ctx, network, devices)
		if err != nil {
			deployErrs <- err
			return
		}

		// Resize the VM's disk.
		diskResizeSpec, err := getDiskDeviceConfigSpec(ctx, n.DiskGiB, devices)
		if err != nil {
			deployErrs <- err
			return
		}

		// Get the extra config data for the VM.
		extraConfig, err := GetExtraConfig(ctx, cfg, *n)
		if err != nil {
			deployErrs <- err
			return
		}

		vmConfigSpec := types.VirtualMachineConfigSpec{
			Name:              n.Name,
			MemoryMB:          int64(n.MemoryMiB),
			NumCPUs:           int32(n.Cores),
			NumCoresPerSocket: int32(n.CoresPerSocket),
			DeviceChange: []types.BaseVirtualDeviceConfigSpec{
				diskResizeSpec,
				netConfigSpec,
			},
			ExtraConfig: extraConfig,
		}

		task, err := n.vm.Reconfigure(ctx, vmConfigSpec)
		if err != nil {
			deployErrs <- err
			return
		}
		result, err := task.WaitForResult(ctx, nil)
		if err != nil {
			deployErrs <- err
			return
		}
		if result.Error != nil && result.Error.LocalizedMessage != "" {
			deployErrs <- errors.New(result.Error.LocalizedMessage)
			return
		}
		n.UUID = n.vm.UUID(ctx)
		deployWait.Done()
	}

	// Kick off the deploy operations.
	fmt.Fprint(out, "deploying nodes...")
	deployWait.Add(len(cfg.Nodes.Nodes))
	for i := range cfg.Nodes.Nodes {
		go deployNode(&cfg.Nodes.Nodes[i])
	}
	go func() {
		deployWait.Wait()
		close(deployDone)
	}()

	// Block until the deploy operations complete.
	if err := wait(out, deployDone, deployErrs); err != nil {
		return &cfg, err
	}

	var (
		powerWait sync.WaitGroup
		powerDone = make(chan struct{})
		powerErrs = make(chan error)
	)

	powerOnNode := func(n *SingleNodeConfig) {
		task, err := n.vm.PowerOn(ctx)
		if err != nil {
			powerErrs <- err
			return
		}
		result, err := task.WaitForResult(ctx, nil)
		if err != nil {
			powerErrs <- err
			return
		}
		if result.Error != nil && result.Error.LocalizedMessage != "" {
			powerErrs <- errors.New(result.Error.LocalizedMessage)
			return
		}
		powerWait.Done()
	}

	// Kick off the power operations.
	fmt.Fprint(out, "powering on nodes...")
	powerWait.Add(len(cfg.Nodes.Nodes))
	for i := range cfg.Nodes.Nodes {
		go powerOnNode(&cfg.Nodes.Nodes[i])
	}
	go func() {
		powerWait.Wait()
		close(powerDone)
	}()

	// Block until the power operations complete.
	if err := wait(out, powerDone, powerErrs); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

func wait(out io.Writer, done chan struct{}, errs chan error) error {
	for {
		select {
		case err := <-errs:
			fmt.Fprintln(out, "failed!")
			return err
		case <-done:
			fmt.Fprintln(out, "success!")
			return nil
		default:
			fmt.Fprint(out, ".")
			time.Sleep(3 * time.Second)
		}
	}
}

// Down destroys a deployed Kubernetes cluster and its associated resources.
// The io.Writer is used to record progress. Use ioutil.Discard or nil to
// ignore progress.
func Down(
	ctx context.Context,
	out io.Writer,
	state Config) error {

	if out == nil {
		out = ioutil.Discard
	}

	// Get a client for the vSphere endpoint.
	client, err := NewVSphereClient(ctx, state.VSphere)
	if err != nil {
		return err
	}

	// Find the objects needed to destroy the VMs.
	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(
		ctx, state.VSphere.Datacenter)
	if err != nil {
		return err
	}
	finder.SetDatacenter(datacenter)

	iu := false
	searchIndex := object.NewSearchIndex(client.Client)

	var (
		powerWait sync.WaitGroup
		powerDone = make(chan struct{})
		powerErrs = make(chan error)
	)

	powerOffNode := func(n *SingleNodeConfig) {
		defer powerWait.Done()
		vmRef, err := searchIndex.FindByUuid(ctx, datacenter, n.UUID, true, &iu)
		if err != nil {
			return
		}
		if vmRef == nil {
			return
		}
		n.vm = object.NewVirtualMachine(client.Client, vmRef.Reference())
		task, err := n.vm.PowerOff(ctx)
		if err != nil {
			powerErrs <- err
			return
		}
		result, err := task.WaitForResult(ctx, nil)
		if err != nil {
			powerErrs <- err
			return
		}
		if result.Error != nil && result.Error.LocalizedMessage != "" {
			powerErrs <- errors.New(result.Error.LocalizedMessage)
			return
		}
	}

	// Kick off the power operations.
	fmt.Fprint(out, "powering off nodes...")
	powerWait.Add(len(state.Nodes.Nodes))
	for i := range state.Nodes.Nodes {
		go powerOffNode(&state.Nodes.Nodes[i])
	}
	go func() {
		powerWait.Wait()
		close(powerDone)
	}()

	// Block until the power operations complete.
	if err := wait(out, powerDone, powerErrs); err != nil {
		return err
	}

	var (
		destroyWait sync.WaitGroup
		destroyDone = make(chan struct{})
		destroyErrs = make(chan error)
	)

	destroyNode := func(n *SingleNodeConfig) {
		defer destroyWait.Done()
		if n.vm == nil {
			return
		}
		task, err := n.vm.Destroy(ctx)
		if err != nil {
			destroyErrs <- err
			return
		}
		result, err := task.WaitForResult(ctx, nil)
		if err != nil {
			destroyErrs <- err
			return
		}
		if result.Error != nil && result.Error.LocalizedMessage != "" {
			destroyErrs <- errors.New(result.Error.LocalizedMessage)
			return
		}
	}

	// Kick off the destroy operations.
	fmt.Fprint(out, "destroying nodes...")
	destroyWait.Add(len(state.Nodes.Nodes))
	for i := range state.Nodes.Nodes {
		go destroyNode(&state.Nodes.Nodes[i])
	}
	go func() {
		destroyWait.Wait()
		close(destroyDone)
	}()

	// Block until the power operations complete.
	if err := wait(out, destroyDone, destroyErrs); err != nil {
		return err
	}

	return nil
}
