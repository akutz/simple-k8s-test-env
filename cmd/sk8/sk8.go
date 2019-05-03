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

package sk8

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"vmware.io/sk8/cmd/sk8/builds"
	"vmware.io/sk8/cmd/sk8/cluster"
	"vmware.io/sk8/cmd/sk8/config"
	"vmware.io/sk8/cmd/sk8/version"
	"vmware.io/sk8/pkg/util"
)

const defaultLevel = log.WarnLevel

// Flags for the sk8 command
type Flags struct {
	LogLevel string
}

// NewCommand returns a new cobra.Command implementing the root command for sk8
func NewCommand() *cobra.Command {
	flags := &Flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "sk8",
		Short: "sk8 is a tool for managing Kubernetes clusters on vSphere",
		Long:  "sk8 is a tool for managing Kubernetes clusters on vSphere",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, cmd, args)
		},
		SilenceUsage: true,
		Version:      version.Version,
	}
	cmd.PersistentFlags().StringVar(
		&flags.LogLevel,
		"log-level",
		defaultLevel.String(),
		"the log level",
	)
	// Add the top-level commands.
	cmd.AddCommand(builds.NewCommand())
	cmd.AddCommand(cluster.NewCommand())
	cmd.AddCommand(config.NewCommand())
	cmd.AddCommand(version.NewCommand())
	return cmd
}

func runE(flags *Flags, cmd *cobra.Command, args []string) error {
	level := defaultLevel
	parsed, err := log.ParseLevel(flags.LogLevel)
	if err != nil {
		log.Warnf(
			"Invalid log level '%s', defaulting to '%s'",
			flags.LogLevel,
			level)
	} else {
		level = parsed
	}
	log.SetLevel(level)
	return nil
}

// Run runs the `sk8` root command
func Run() error {
	return NewCommand().Execute()
}

// Main wraps Run and sets the log formatter.
func Main() {
	// Explicitly log to stdout.
	log.SetOutput(os.Stdout)
	// Make the default log formatter's timestamps a little more useful.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
		ForceColors:     util.IsTerminal(log.StandardLogger().Out),
	})
	if err := Run(); err != nil {
		os.Exit(1)
	}
}
