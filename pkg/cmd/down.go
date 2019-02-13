package cmd // import "vmw.io/sk8/cmd"

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vmw.io/sk8/app"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Destroy a Kubernetes cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if stateFilePath == "" {
			return fmt.Errorf("-state is required")
		}
		f, err := os.Open(stateFilePath)
		if err != nil {
			return fmt.Errorf(
				"error opening state file %s: %v", stateFilePath, err)
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		if err := dec.Decode(&config.Config); err != nil {
			return fmt.Errorf(
				"error loading state file %s: %v", stateFilePath, err)
		}
		if err := config.Build(ctx); err != nil {
			return fmt.Errorf("error building config: %v", err)
		}
		return app.Down(ctx, os.Stdout, config.Config)
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
	downCmd.Flags().StringVar(
		&stateFilePath,
		"state",
		"",
		"The path to a state file.")
}
