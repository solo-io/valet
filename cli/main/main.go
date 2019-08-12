package main

import (
	"github.com/solo-io/valet/cli/cmd"
	"os"
)

func main() {
	app := cmd.SoloCli("todo")
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}