package cmd

import (
	"context"
	"github.com/solo-io/valet/cli/cmd/config"
	"github.com/solo-io/valet/cli/cmd/ensure"
	"github.com/solo-io/valet/cli/cmd/teardown"
	"github.com/solo-io/valet/cli/options"

	"github.com/solo-io/go-utils/cliutils"

	"github.com/spf13/cobra"
)

func App(version string, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {

	app := &cobra.Command{
		Use:     "valet",
		Short:   "CLI for ensuring the state of clusters, solo products, and demos",
		Version: version,
	}

	// Complete additional passed in setup
	cliutils.ApplyOptions(app, optionsFunc)

	return app
}

func ValetCli(version string) *cobra.Command {
	opts := &options.Options{
		Top: options.Top{
			Ctx: context.Background(),
		},
	}

	optionsFunc := func(app *cobra.Command) {
		app.SuggestionsMinimumDistance = 1
		app.AddCommand(
			ensure.Ensure(opts),
			teardown.Teardown(opts),
			config.Config(opts),
		)
	}

	return App(version, optionsFunc)
}
