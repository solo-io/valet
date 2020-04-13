package main

import (
	"os"

	"github.com/solo-io/valet/pkg/cli/cmd"
	"github.com/solo-io/valet/pkg/version"
)

func main() {
	app := cmd.ValetCli(version.Version)
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
