package cmd // import "vmw.io/sk8/cmd"

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vmw.io/sk8/app"
)

var (
	config struct {
		app.Config
		filePath string
	}
	ctx           = context.Background()
	stateFilePath string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sk8",
	Short: "A turn-key solution for deploying Kubernetes on VMware vSphere",
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen
// once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(
		&config.VSphere.ServerURL,
		"vsphere-url",
		"",
		"A VMware vSphere endpoint, ex. SCHEME://SERVER[:PORT]/sdk. May be set with SK8_VSPHERE_URL.")

	rootCmd.PersistentFlags().StringVar(
		&config.VSphere.Username,
		"vsphere-username",
		"",
		"The username used to access vsphere-url. May be set with SK8_VSPHERE_USERNAME.")

	rootCmd.PersistentFlags().StringVar(
		&config.VSphere.Password,
		"vsphere-password",
		"",
		"The password used to access vsphere-url. May be set with SK8_VSPHERE_PASSWORD.")

	rootCmd.PersistentFlags().BoolVar(
		&config.VSphere.Insecure,
		"vsphere-insecure",
		false,
		"A flag indicating whether or not to verify a peer's TLS certificate. May be set with SK8_VSPHERE_INSECURE.")
}

func initConfig() {
	if config.filePath != "" {
		f, err := os.Open(config.filePath)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"error reading config %s: %v\n",
				config.filePath, err)
			os.Exit(1)
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		if err := dec.Decode(&config.Config); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"error loading config %s: %v\n",
				config.filePath, err)
			os.Exit(1)
		}
	}
}
