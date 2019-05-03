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

package builds

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	apimv "k8s.io/apimachinery/pkg/util/version"
)

const (
	ciBuildURI      = "https://storage.googleapis.com/kubernetes-release-dev"
	releaseBuildURI = "https://storage.googleapis.com/kubernetes-release"
)

// Resolve returns the URI and version of a Kubernetes build for the given
// build ID.
//
// The given buildID may be defined three ways:
//
//   1. A local filesystem path containing the files listed above in a flat
//      structure.
//
//   2. An HTTP address that contains the files listed above in a structure
//      that adheres to the layout of the public GCS buckets kubernetes-release
//      and kubernetes-release-dev. For example, if buildID is set to
//      https://k8s.ci/v1.13 then the following URI should be valid:
//
//        * https://k8s.ci/v1.13/kubernetes.tar.gz
//
//   3. A valid semantic version for a released version of Kubernetes or
//      begins with "ci/" or "release/". If the string matches any of these
//      then the value is presumed to be a CI or release build hosted on one
//      of the public GCS buckets.
//
//      This option also supports values ended with ".txt", ex. "ci/latest.txt".
//      In fact, all values beginning with "ci/" or "release/" are first checked
//      to see if there's a matching txt file for that value. For example,
//      if "ci/latest" is provided then before assuming that is a directory,
//      "ci/latest.txt" is queried.
//
// In order to prevent the second two options from conflicting with the first,
// a local file path, using the prefix "file://" explicitly indicates the given
// buildID is a local file path.
func Resolve(buildID string) (string, string, error) {

	var uri string

	if ok, _ := regexp.MatchString(`(?i)^https?:`, buildID); ok {
		uri = buildID
	} else if _, err := apimv.ParseGeneric(buildID); err == nil {
		uri = fmt.Sprintf("%s/release/%s", releaseBuildURI, buildID)
	} else if strings.HasPrefix(buildID, "ci/") {
		uri2, err := resolveBuildURI(buildID, true)
		if err != nil {
			return "", "", errors.Wrapf(
				err, "error resolving ci build URI %s", buildID)
		}
		uri = uri2
	} else if strings.HasPrefix(buildID, "release/") {
		uri2, err := resolveBuildURI(buildID, false)
		if err != nil {
			return "", "", errors.Wrapf(
				err, "error resolving release build URI %s", buildID)
		}
		uri = uri2
	} else {
		return "", "", errors.Errorf("invalid buildID %s", buildID)
	}

	kubeTarballURI := uri + "/kubernetes.tar.gz"
	_, r, err := httpGet(kubeTarballURI)
	if err != nil {
		return "", "", errors.Wrapf(err, "error reading %s", kubeTarballURI)
	}
	defer r.Close()

	version, err := readVersionFromKubeTarball(r)
	if err != nil {
		return "", "", errors.Wrapf(
			err, "error reading version from %s", kubeTarballURI)
	}

	return uri, version, nil
}

func resolveBuildURI(buildID string, ciBuild bool) (string, error) {
	var uri string
	if ciBuild {
		uri = fmt.Sprintf("%s/%s", ciBuildURI, buildID)
	} else {
		uri = fmt.Sprintf("%s/%s", releaseBuildURI, buildID)
	}

	// If the URI doesn't end with ".txt" then see if the URI is already valid.
	if !strings.HasSuffix(uri, ".txt") {

		// If there is a kubernetes tarball available at the root of the URI
		// then it is already a valid URI.
		if ok, _ := httpHeadOK(uri + "/kubernetes.tar.gz"); ok {
			return uri, nil
		}

		// The URI wasn't valid, so add ".txt" to the end and let's see if the
		// URI points to a valid build.
		uri = uri + ".txt"
	}

	// Do an HTTP GET and read the version from the txt file.
	_, r, err := httpGet(uri)
	if err != nil {
		return "", errors.Wrapf(err, "invalid URI: %s", uri)
	}
	defer r.Close()
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return "", errors.Wrapf(err, "error reading version from %s", uri)
	}
	version := url.PathEscape(string(bytes.TrimSpace(buf)))

	// Format the URI based on whether or not it's a CI or release build.
	if ciBuild {
		return fmt.Sprintf("%s/ci/%s", ciBuildURI, version), nil
	}
	return fmt.Sprintf("%s/release/%s", releaseBuildURI, version), nil
}

// readVersionFromKubeTarball reads the version file from the
// kubernetes.tar.gz file at the given uri.
func readVersionFromKubeTarball(r io.Reader) (string, error) {
	// Create a gzip reader to process the source reader.
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return "", errors.Wrap(err, "error getting gzip reader")
	}

	// Create a tar reader to process the gzip reader.
	tarReader := tar.NewReader(gzipReader)

	// Iterate until able to read the version file.
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return "", errors.New(
					"failed to find version in kube tarball")
			}
			return "", errors.Wrap(
				err, "error iterating kube tarball")
		}
		if header.Name == "kubernetes/version" {
			version, err := ioutil.ReadAll(tarReader)
			if err != nil {
				return "", errors.Wrap(
					err, "error reading version from kube tarball")
			}
			return string(bytes.TrimSpace(version)), nil
		}
	}
}

func httpHeadOK(uri string) (bool, error) {
	resp, err := http.Head(uri)
	if err != nil {
		return false, errors.Wrapf(err, "HTTP HEAD %s failed", uri)
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("HTTP HEAD %s failed: %s", uri, resp.Status)
	}
	return true, nil
}

func httpGet(uri string) (int64, io.ReadCloser, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return 0, nil, errors.Wrapf(err, "HTTP GET %s failed", uri)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil, errors.Errorf("HTTP GET %s failed: %s", uri, resp.Status)
	}
	return resp.ContentLength, resp.Body, nil
}
