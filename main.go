package main

import (
	"os"

	"github.com/solo-io/valet/cli/cmd"
	"github.com/solo-io/valet/cli/version"
)

func main() {
	app := cmd.ValetCli(version.Version)
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
