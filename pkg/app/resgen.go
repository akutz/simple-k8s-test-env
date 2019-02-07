//go:generate go run ../resgen/main.go -project-root ../..

package app // import "github.com/vmware/sk8/pkg/app"

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	sk8ScriptRes string
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
