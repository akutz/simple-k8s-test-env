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

package cluster

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"regexp"

	"github.com/pkg/errors"
)

var (
	// ErrNotFound is returned if a file is not found.
	ErrNotFound = errors.New("not found")

	// dataDir is the user-scoped directory where
	dataDir string
)

func init() {
	if u, err := user.Current(); err == nil {
		dataDir = path.Join(u.HomeDir, ".sk8")
	}
}

// DataDir returns the path to the local data directory for the current
// user.
// An empty string is returned if an error occurred while discovering the
// data directory.
func DataDir() string {
	return dataDir
}

// Default returns a Cluster object if there is exactly one name
// returned from a call to List. Otherwise ErrNotFound is returned.
func Default() (*Cluster, error) {
	configFiles := listConfigFiles()
	if len(configFiles) != 1 {
		return nil, ErrNotFound
	}
	return ReadFromFile(configFiles[0])
}

// DefaultName returns a cluster name there is exactly one name
// returned from a call to List. Otherwise an empty string is returned.
func DefaultName() string {
	names := ListNames()
	if len(names) != 1 {
		return ""
	}
	return names[0]
}

// FilePath returns a path rooted in the cluster's data directory.
// An empty string is returned if an error occurred while discovering the
// data directory.
func FilePath(clusterName string, paths ...string) string {
	if dataDir == "" || ValidateName(clusterName) != nil {
		return ""
	}
	return path.Join(append([]string{dataDir, clusterName}, paths...)...)
}

// List returns the known clusters.
func List() ([]*Cluster, error) {
	configFiles := listConfigFiles()
	clusters := make([]*Cluster, len(configFiles))
	for i, f := range configFiles {
		c, err := ReadFromFile(f)
		if err != nil {
			return nil, err
		}
		clusters[i] = c
	}
	return clusters, nil
}

// ListNames returns the names of the known clusters.
func ListNames() []string {
	configFiles := listConfigFiles()
	names := make([]string, len(configFiles))
	for i := range configFiles {
		names[i] = path.Base(path.Dir(configFiles[i]))
	}
	return names
}

func listConfigFiles() []string {
	if dataDir == "" {
		return nil
	}
	f, err := os.Open(dataDir)
	if err != nil {
		return nil
	}
	defer f.Close()
	files, err := f.Readdir(0)
	if err != nil {
		return nil
	}
	configFiles := []string{}
	for _, info := range files {
		fileName := info.Name()
		if info.IsDir() && ValidateName(fileName) == nil {
			confFile := path.Join(dataDir, fileName, "sk8.conf")
			if _, err := os.Stat(confFile); err == nil {
				configFiles = append(configFiles, confFile)
			}
		}
	}
	return configFiles
}

// Load returns the Cluster object for the given clusterName.
// If no cluster is found then ErrNotFound is returned.
func Load(clusterName string) (*Cluster, error) {
	confFile := FilePath(clusterName, "sk8.conf")
	if _, err := os.Stat(confFile); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ReadFromFile(confFile)
}

// Must is a wrapper for the package functions that return a *Cluster
// and an error. Using Must causes the program to panic if the error is
// non-nil.
func Must(obj *Cluster, val error) *Cluster {
	if val != nil {
		panic(val)
	}
	return obj
}

// NewName generates a new cluster name.
func NewName() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	h := sha1.New()
	h.Write(buf)
	id := fmt.Sprintf("%x", h.Sum(nil))
	return "sk8-" + id[:7]
}

// ReadFrom reads a sk8 cluster configuration from the given io.Reader.
func ReadFrom(r io.Reader) (*Cluster, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var obj Cluster
	if err := obj.UnmarshalText(buf); err != nil {
		return nil, err
	}
	return &obj, nil
}

// ReadFromFile reads a sk8 cluster configuration file.
func ReadFromFile(path string) (*Cluster, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadFrom(f)
}

// ValidateName returns a nil error to indicate the given val is valid.
func ValidateName(val string) error {
	ok, err := regexp.MatchString(`^sk8-\w{7}$`, val)
	if err != nil {
		return errors.Wrapf(err, "invalid cluster name %q", val)
	}
	if !ok {
		return errors.Errorf("invalid cluster name %q", val)
	}
	return nil
}

// WithStdDefaults updates the Cluster's ClusterProviderConfig object and all of
// the MachineProviderConfig objects using the sk8 default configuration data
// from the following files, if they exist:
//   1. /etc/sk8/sk8.conf
//   2. $HOME/.sk8/sk8.conf
func WithStdDefaults(c *Cluster) (*Cluster, error) {
	var paths []string

	// System conf
	{
		p := path.Join("/etc", "sk8", "sk8.conf")
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, p)
		}
	}

	// User conf
	{
		if dataDir != "" {
			p := path.Join(dataDir, "sk8.conf")
			if _, err := os.Stat(p); err == nil {
				paths = append(paths, p)
			}
		}
	}

	if len(paths) > 0 {
		return c.WithDefaults(paths...)
	}

	return c, nil
}
