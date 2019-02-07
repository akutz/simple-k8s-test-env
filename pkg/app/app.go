package app // import "github.com/vmware/sk8/pkg/app"

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"

	"github.com/vmware/govmomi/object"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/sk8/pkg/config"
)

// Up deploys a Kubernetes cluster using the provided configuration and
// returns the result of the operation. The io.Writer is used to record
// progress. Use ioutil.Discard or nil to ignore progress.
func Up(
	ctx context.Context,
	out io.Writer,
	cfg config.Config) (*State, error) {

	if out == nil {
		out = ioutil.Discard
	}

	if err := ValidateConfig(ctx, &cfg); err != nil {
		return nil, err
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
	item, err := libClient.UploadOVA(
		ctx,
		library.Library{
			Name: "sk8",
			Storage: []library.StorageBackings{
				library.StorageBackings{
					Type:        "DATASTORE",
					DatastoreID: datastore.Reference().Value,
				},
			},
			Type: "LOCAL",
		},
		library.Item{
			Name: "photon2-cloud-init",
			Type: "ovf",
		},
		"https://s3-us-west-2.amazonaws.com/cnx.vmware/photon2-cloud-init.ova",
		false)
	if err != nil {
		return nil, err
	}
	_ = item

	vm2clone, err := finder.VirtualMachine(ctx, cfg.VSphere.Template)
	if err != nil {
		return nil, err
	}

	// Get the device list of the VM being cloned.
	devices, err := vm2clone.Device(ctx)
	if err != nil {
		return nil, err
	}

	// A list of device config specs for the cloned VM.
	configSpecs := []types.BaseVirtualDeviceConfigSpec{}

	// Search for the first network card of the source and update
	// the config spec's network device if necessary.
	netConfigSpec, err := getNetworkDeviceConfigSpec(ctx, network, devices)
	if err != nil {
		return nil, err
	}
	configSpecs = append(configSpecs, netConfigSpec)

	datastoreRef := datastore.Reference()
	folderRef := folder.Reference()
	poolRef := pool.Reference()

	relocateSpec := types.VirtualMachineRelocateSpec{
		Datastore:    &datastoreRef,
		DeviceChange: configSpecs,
		Folder:       &folderRef,
		Pool:         &poolRef,
	}

	// Create the state in order to push node info into it.
	state := &State{}

	fmt.Fprintf(out, "creating nodes:\n")

	// Iterate over all of the nodes.
	var (
		controlPlaneNodeIndex int
		workerNodeIndex       int
	)
	for _, n := range cfg.Nodes {

		// Generate the host name, host FQDN, and VM name.
		var (
			hostName string
			hostFQDN string
			nodeType string
			vmName   string
		)
		switch n.Type {
		case config.ControlPlaneNode, config.ControlPlaneWorkerNode:
			if n.Type == config.ControlPlaneNode {
				nodeType = "controller"
			} else {
				nodeType = "both"
			}
			controlPlaneNodeIndex++
			hostName = fmt.Sprintf("c%02d", controlPlaneNodeIndex)
		case config.WorkerNode:
			nodeType = "worker"
			workerNodeIndex++
			hostName = fmt.Sprintf("w%02d", workerNodeIndex)
		}
		hostFQDN = fmt.Sprintf("%s.%s", hostName, cfg.Network.DomainFQDN)
		vmName = fmt.Sprintf(
			"%s.%s", hostName, strings.Split(cfg.Network.DomainFQDN, ".")[0])

		// Get the extra config data for the VM.
		extraConfig, err := GetExtraConfig(ctx, cfg, hostFQDN, nodeType)
		if err != nil {
			return state, err
		}

		diskResizeSpec, err := getDiskDeviceConfigSpec(ctx, n.DiskGiB, devices)
		if err != nil {
			return state, err
		}

		vmConfigSpec := types.VirtualMachineConfigSpec{
			Name:              vmName,
			MemoryMB:          int64(n.MemoryMiB),
			NumCPUs:           int32(n.Cores),
			NumCoresPerSocket: int32(n.CoresPerSocket),
			DeviceChange: []types.BaseVirtualDeviceConfigSpec{
				diskResizeSpec,
			},
			ExtraConfig: extraConfig,
		}

		cloneSpec := types.VirtualMachineCloneSpec{
			PowerOn:  false,
			Template: false,
			Location: relocateSpec,
			Config:   &vmConfigSpec,
		}

		fmt.Fprintf(out, "  %s...", vmName)

		cloneTask, err := vm2clone.Clone(ctx, folder, vmName, cloneSpec)
		if err != nil {
			fmt.Fprintf(out, "failed!\n")
			return state, err
		}

		cloneResult, err := cloneTask.WaitForResult(ctx, nil)
		if err != nil {
			fmt.Fprintf(out, "failed!\n")
			return state, err
		}

		if cloneResult.Error != nil &&
			cloneResult.Error.LocalizedMessage != "" {
			fmt.Fprintf(out, "failed!\n")
			return state, errors.New(cloneResult.Error.LocalizedMessage)
		}

		fmt.Fprintf(out, "success!\n")

		vm := object.NewVirtualMachine(client.Client,
			cloneResult.Result.(types.ManagedObjectReference))
		vmUUID := vm.UUID(ctx)

		state.Nodes = append(state.Nodes, NodeInfo{
			Type:     n.Type,
			HostFQDN: hostFQDN,
			HostName: hostName,
			UUID:     vmUUID,
			vm:       vm,
			vmName:   vmName,
		})

		if len(state.ClusterID) == 0 {
			state.ClusterID = vmUUID
		}
	}

	fmt.Fprintf(out, "powering on nodes:\n")
	for _, n := range state.Nodes {
		fmt.Fprintf(out, "  %s...", n.vmName)
		powerOnTask, err := n.vm.PowerOn(ctx)
		if err != nil {
			fmt.Fprintf(out, "failed!\n")
			return state, err
		}
		powerOnResult, err := powerOnTask.WaitForResult(ctx, nil)
		if err != nil {
			fmt.Fprintf(out, "failed!\n")
			return state, err
		}
		if powerOnResult.Error != nil &&
			powerOnResult.Error.LocalizedMessage != "" {
			fmt.Fprintf(out, "failed!\n")
			return state, errors.New(powerOnResult.Error.LocalizedMessage)
		}
		fmt.Fprintf(out, "success!\n")
	}

	return state, nil
}

// Down destroys a deployed Kubernetes cluster and its associated resources.
// The io.Writer is used to record progress. Use ioutil.Discard or nil to
// ignore progress.
func Down(
	ctx context.Context,
	out io.Writer,
	state State) error {

	return nil
}
