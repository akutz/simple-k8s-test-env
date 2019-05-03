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

package util

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Base64GzipReader gzips an IO stream and returns the data as a
// base64-encoded string.
func Base64GzipReader(
	r io.Reader,
	name string,
	modTime time.Time) (string, error) {

	gzipData := &bytes.Buffer{}
	gzipWriter, err := gzip.NewWriterLevel(gzipData, gzip.BestCompression)
	if err != nil {
		return "", err
	}

	if len(name) > 0 {
		gzipWriter.Name = name
	}
	if !modTime.IsZero() {
		gzipWriter.ModTime = modTime
	}

	if _, err := io.Copy(gzipWriter, r); err != nil {
		return "", err
	}
	if err := gzipWriter.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipData.Bytes()), nil
}

// Base64GzipBytes gzips a buffer and returns the data as a
// base64-encoded string.
func Base64GzipBytes(data []byte) (string, error) {
	return Base64GzipReader(bytes.NewReader(data), "", time.Time{})
}

// Base64GzipFile gzips a file and returns the data as a base64-encoded string.
func Base64GzipFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	return Base64GzipReader(file, filepath.Base(path), fileInfo.ModTime())
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
