package app // import "github.com/vmware/sk8/pkg/app"

import "os"

// FileExists returns a flag indicating whether a provided file path exists.
func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return true
	}
	return false
}
