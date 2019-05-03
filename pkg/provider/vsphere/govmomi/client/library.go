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
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/library/finder"
	"github.com/vmware/govmomi/vapi/vcenter"
	"github.com/vmware/govmomi/vim25/types"

	"vmware.io/sk8/pkg/provider/vsphere/config"
)

// EnsureLibraryOVA ensures an OVA is present in the content library and
// returns the content library item ID.
func (c *Client) EnsureLibraryOVA(
	ctx context.Context,
	cfg config.ImportOVAConfig) (string, error) {

	libManager := library.NewManager(c.Rest)
	libFinder := finder.NewFinder(libManager)
	importTarget := strings.Split(cfg.Target, "/")
	if importTarget[0] == "" {
		importTarget = importTarget[1:]
	}
	libName := importTarget[0]
	libPath := importTarget[0]
	libItemName := importTarget[1]
	libItemPath := cfg.Target

	var libID string
	var libItemID string

	log.WithFields(log.Fields{
		"source": cfg.Source,
		"target": cfg.Target,
	}).Debug("ensuring content library OVA")

	// Get or create the content library.
	libs, err := libFinder.Find(ctx, libPath)
	if err != nil {
		return "", errors.Wrapf(
			err, "error finding content library %q", libPath)
	}
	if len(libs) > 0 {
		libID = libs[0].GetID()
		log.WithFields(log.Fields{
			"id":   libID,
			"path": libPath,
		}).Debug("discovered content library")
	} else {
		libID2, err := libManager.CreateLibrary(ctx, library.Library{
			Name: libName,
			Storage: []library.StorageBackings{
				{
					Type:        "DATASTORE",
					DatastoreID: c.Datastore.Reference().Value,
				},
			},
			Type: "LOCAL",
		})
		if err != nil {
			return "", errors.Wrapf(
				err, "error creating content library %q", libName)
		}
		libID = libID2
		log.WithFields(log.Fields{
			"id":   libID,
			"path": libPath,
		}).Debug("created content library")
	}

	// Get or create the content library item.
	libItems, err := libFinder.Find(ctx, libItemPath)
	if err != nil {
		return "", errors.Wrapf(
			err, "error finding content library item %q", libItemPath)
	}
	if len(libItems) > 0 {
		libItemID = libItems[0].GetID()
		log.WithFields(log.Fields{
			"id":   libItemID,
			"path": libItemPath,
		}).Debug("discovered content library item")
		return libItemID, nil
	}

	libItemID2, err := libManager.CreateLibraryItem(ctx, library.Item{
		Name:      libItemName,
		LibraryID: libID,
		Type:      "ovf",
	})
	if err != nil {
		return "", errors.Wrapf(
			err, "error creating content library item %q", libItemPath)
	}
	libItemID = libItemID2

	// Create an UpdateSession used to upload the remote file to the
	// content library item.
	sessionSpec := library.UpdateSession{LibraryItemID: libItemID}
	sessionID, err := libManager.CreateLibraryItemUpdateSession(
		ctx, sessionSpec)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"error creating content library item update session for %q",
			libItemPath)
	}

	// Get the spec used to upload the remote file.
	if _, err := libManager.AddLibraryItemFileFromURI(
		ctx, sessionID,
		"file.ova",
		cfg.Source); err != nil {
		return "", errors.Wrapf(
			err,
			"error adding %q to content library item %q",
			cfg.Source,
			libItemPath)
	}

	// Wait until the upload operation is complete to return.
	if err := libManager.WaitOnLibraryItemUpdateSession(
		ctx, sessionID, time.Duration(3)*time.Second, nil); err != nil {
		return "", errors.Wrapf(
			err,
			"error waiting on %q to upload to content library item %q",
			cfg.Source,
			libItemPath)
	}

	log.WithFields(log.Fields{
		"id":   libItemID,
		"path": libItemPath,
	}).Debug("created content library item")
	return libItemID, nil
}

// DeployOVF deploys a VM from an OVF that resides in the content library.
func (c *Client) DeployOVF(
	ctx context.Context,
	itemID string,
	spec vcenter.Deploy) (*object.VirtualMachine, error) {

	echan := make(chan error)
	rchan := make(chan vcenter.DeployedResourceID)

	go func() {
		vcm := vcenter.NewManager(c.Rest)
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
			return nil, err
		case ref := <-rchan:
			return object.NewVirtualMachine(
				c.Vim25,
				types.ManagedObjectReference{
					Value: ref.ID,
					Type:  ref.Type,
				}), nil
		default:
			time.Sleep(3 * time.Second)
		}
	}
}
