package cmd

import (
	"context"
	"github.com/solo-io/valet/pkg/cli/cmd/config"
	gen_docs "github.com/solo-io/valet/pkg/cli/cmd/gen-docs"
	"github.com/solo-io/valet/pkg/cli/cmd/run"
	"github.com/solo-io/valet/pkg/cli/options"

	"github.com/solo-io/go-utils/cliutils"

	"github.com/spf13/cobra"
)

func App(version string, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {

	app := &cobra.Command{
		Use:     "valet",
		Short:   "CLI for running valet workflows",
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
		app.PersistentFlags().StringVarP(&opts.Config.GlobalConfigPath, "global-config-path", "", "", "alternate location for global config (default $HOME/.valet/global.yaml)")
		app.AddCommand(
			run.Run(opts),
			config.Config(opts),
			gen_docs.GenDocs(opts),
		)
	}

	return App(version, optionsFunc)
}
