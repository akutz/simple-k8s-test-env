package app // import "github.com/vmware/sk8/pkg/app"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/vmware/govmomi/vapi/vcenter"

	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"
)

// LibraryClient is a client for the vSphere content library.
type LibraryClient struct {
	*rest.Client
	m *library.Manager
	w io.Writer
	u string
}

// NewLibraryClient returns a new client for the vSphere content library.
func NewLibraryClient(
	client *rest.Client,
	out io.Writer) *LibraryClient {

	if out == nil {
		out = ioutil.Discard
	}
	return &LibraryClient{
		Client: client,
		m:      library.NewManager(client),
		w:      out,
		u:      strings.Replace(client.URL().String(), "/sdk", "", 1),
	}
}

// CreateLibraryIfNotExists creates a new library with the provided
// name or creates a new item if a match is not found.
func (c *LibraryClient) CreateLibraryIfNotExists(
	ctx context.Context,
	lib library.Library) (*library.Library, error) {

	fmt.Fprintf(c.w, "finding library...")
	if lib, _ := c.m.GetLibraryByName(ctx, lib.Name); lib != nil {
		fmt.Fprintf(c.w, "success!\n")
		return lib, nil
	}
	fmt.Fprintf(c.w, "not found\n")

	fmt.Fprintf(c.w, "creating library...")
	libID, err := c.m.CreateLibrary(ctx, lib)
	if err != nil {
		fmt.Fprintf(c.w, "failed!\n")
		return nil, err
	}
	fmt.Fprintf(c.w, "success!\n")
	lib.ID = libID
	return &lib, nil
}

// CreateLibraryItemIfNotExists creates a new library item with the provided
// name and type or creates a new item if a match is not found.
func (c *LibraryClient) CreateLibraryItemIfNotExists(
	ctx context.Context,
	item library.Item) (result *library.Item, err error) {

	fmt.Fprintf(c.w, "finding library item...")
	foundItemIDs, err := c.m.FindLibraryItems(
		ctx, library.FindLibraryItemsRequest{
			Name:      item.Name,
			Type:      item.Type,
			LibraryID: item.LibraryID,
		})
	if err != nil {
		fmt.Fprintf(c.w, "failed\n")
		return nil, err
	}
	if len(foundItemIDs) > 0 {
		item2, err := c.m.GetLibraryItem(ctx, foundItemIDs[0])
		if err != nil {
			fmt.Fprintf(c.w, "failed\n")
			return nil, err
		}
		fmt.Fprintf(c.w, "success\n")
		return item2, nil
	}
	fmt.Fprintf(c.w, "not found\n")

	fmt.Fprintf(c.w, "creating library item...")
	defer func() {
		if result != nil {
			fmt.Fprintf(c.w, "success!\n")
		} else {
			fmt.Fprintf(c.w, "failed!\n")
		}
	}()

	itemID, err := c.m.CreateLibraryItem(ctx, item)
	if err != nil {
		return nil, err
	}
	item.ID = itemID

	return &item, nil
}

// DeployOVF deploys a VM from an OVF that resides in the content library.
func (c *LibraryClient) DeployOVF(
	ctx context.Context,
	itemID string,
	spec vcenter.Deploy) (*types.ManagedObjectReference, error) {

	fmt.Fprintf(c.w, "deploying ovf...")
	echan := make(chan error)
	rchan := make(chan vcenter.DeployedResourceID)

	go func() {
		vcm := vcenter.NewManager(c.Client)
		d, err := vcm.DeployLibraryItem(ctx, itemID, spec)
		if err != nil {
			echan <- err
		} else if !d.Succeeded {
			echan <- fmt.Errorf("%+v", d.Error)
		} else {
			rchan <- d.ResourceID
		}
	}()

	// Block until the deploy operation completes.
	for {
		select {
		case err := <-echan:
			fmt.Fprintf(c.w, "failed!\n")
			return nil, err
		case ref := <-rchan:
			fmt.Fprintf(c.w, "success!\n")
			return &types.ManagedObjectReference{
				Value: ref.ID,
				Type:  ref.Type,
			}, nil
		default:
			fmt.Fprintf(c.w, ".")
			time.Sleep(3 * time.Second)
		}
	}
}

// UpdateInfo is information about an upload as well as the session ID.
type UpdateInfo struct {
	library.UpdateFileInfo
	SessionID string
	Length    int64
}

// GetObjectJSONReader returns an io.Reader into which the provided
// object is encoded as JSON.
func GetObjectJSONReader(i interface{}) (io.Reader, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(i); err != nil {
		return nil, err
	}
	return buf, nil
}

// UploadOVA uploads a remote OVA to a content library.
func (c *LibraryClient) UploadOVA(
	ctx context.Context,
	lib library.Library, item library.Item,
	uri string, overwrite bool) (result *library.Item, err error) {

	// Create or get a content library.
	lib2, err := c.CreateLibraryIfNotExists(ctx, lib)
	if err != nil {
		return nil, err
	}

	// Create or get a content library item.
	item.LibraryID = lib2.ID
	item2, err := c.CreateLibraryItemIfNotExists(ctx, item)
	if err != nil {
		return nil, err
	}

	// Check to see if the OVF file is already present. This indicates
	// the OVA has already been uploaded.
	if !overwrite {
		fmt.Fprintf(c.w, "finding ovf...")
		if file, _ := c.m.GetLibraryItemFile(
			ctx, item2.ID, fmt.Sprintf("%s.ovf", item2.Name)); file != nil {

			fmt.Fprintf(c.w, "success!\n")
			return item2, nil
		}
		fmt.Fprintf(c.w, "not found\n")
	}

	fmt.Fprintf(c.w, "uploading ova...")
	defer func() {
		if err != nil {
			fmt.Fprintf(c.w, "failed!\n")
		} else {
			fmt.Fprintf(c.w, "success!\n")
		}
	}()

	// Create an UpdateSession used to upload the remote file to the
	// content library item.
	sessionSpec := library.UpdateSession{LibraryItemID: item2.ID}
	sessionID, err := c.m.CreateLibraryItemUpdateSession(ctx, sessionSpec)
	if err != nil {
		return nil, err
	}

	// Get the spec used to upload the remote file.
	if _, err := c.m.AddLibraryItemFileFromURI(
		ctx, sessionID,
		fmt.Sprintf("%s.ova", item2.Name), uri); err != nil {
		return nil, err
	}

	// Wait until the upload operation is complete to return.
	if err := c.m.WaitOnLibraryItemUpdateSession(
		ctx, sessionID, time.Duration(3)*time.Second,
		func() { fmt.Fprintf(c.w, ".") }); err != nil {
		return nil, err
	}

	return item2, nil
}
