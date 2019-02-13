package cmd // import "vmw.io/sk8/cmd"

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"vmw.io/sk8/app"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Deploy a Kubernetes cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Build(ctx); err != nil {
			return fmt.Errorf("error building config: %v", err)
		}
		state, err := app.Up(ctx, os.Stdout, config.Config)
		if err != nil {
			return fmt.Errorf("error deploying cluster: %v", err)
		}
		var (
			w           io.Writer = os.Stdout
			isStateFile bool
		)
		if stateFilePath != "" {
			f, err := os.Create(stateFilePath)
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"failed to create state file %s: %v",
					stateFilePath,
					err)
			}
			defer f.Close()
			w = f
			isStateFile = true
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(state); err != nil {
			return fmt.Errorf("failed to save state: %v", err)
		}
		if isStateFile {
			fmt.Fprintf(os.Stdout, "cluster state saved to %s\n", stateFilePath)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVar(
		&config.filePath,
		"config",
		"",
		"The path to the config file used to deploy a cluster.")
	upCmd.Flags().StringVar(
		&stateFilePath,
		"state",
		"",
		"The path to record the result of the operation.")
	upCmd.Flags().StringVar(
		&config.K8s.Version,
		"k8s-version",
		"release/stable",
		"The version of Kubernetes to deploy. May be set with SK8_K8S_VERSION.")
}
