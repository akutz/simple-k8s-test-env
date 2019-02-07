package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"vmw.io/sk8/app"
	"vmw.io/sk8/config"
)

var (
	configFilePath string
	stateFilePath  string
)

func main() {
	flag.StringVar(
		&configFilePath,
		"config",
		"",
		"The config file")
	flag.StringVar(
		&stateFilePath,
		"state",
		"",
		"The state file. If omitted a state file is written to the "+
			"working directory using the name of the provided config "+
			"file with its extension replaced with \"state\".")
	flag.Parse()

	cfg := config.Config{}

	if len(configFilePath) > 0 {
		configFile, err := os.Open(configFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading config file: %v\n", err)
			os.Exit(1)
		}
		defer configFile.Close()

		dec := json.NewDecoder(configFile)
		if err := dec.Decode(&cfg); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing config file: %v\n", err)
			os.Exit(1)
		}
	}

	ctx := context.Background()
	state, err := app.Up(ctx, os.Stdout, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating cluster: %v\n", err)
		os.Exit(1)
	}

	if len(stateFilePath) == 0 {
		stateFilePath = fmt.Sprintf(
			"%s.state", fileNameNoExt(filepath.Base(configFilePath)))
	}
	stateFile, err := os.Create(stateFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating state file: %v\n", err)
		os.Exit(1)
	}
	defer stateFile.Close()
	enc := json.NewEncoder(stateFile)
	enc.SetIndent("", "  ")
	enc.Encode(state)
}

var fileNameNoExtRX = regexp.MustCompile(`^(.+?)(?:\.[^.]+)?$`)

func fileNameNoExt(fileName string) string {
	m := fileNameNoExtRX.FindStringSubmatch(fileName)
	if len(m) < 2 {
		return fileName
	}
	return m[1]
}
