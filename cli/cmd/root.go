package cmd

import (
	"context"
	"github.com/solo-io/kube-cluster/cli/ensure"
	"github.com/solo-io/kube-cluster/cli/options"

	"github.com/solo-io/go-utils/cliutils"

	"github.com/spf13/cobra"
)

func App(version string, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {

	app := &cobra.Command{
		Use:   "soloctl",
		Short: "CLI for Solo products",
		Long: `soloctl is the unified CLI for solo products.
	Find more information at https://solo.io`,
		Version: version,
	}

	// Complete additional passed in setup
	cliutils.ApplyOptions(app, optionsFunc)

	return app
}

func SoloCli(version string) *cobra.Command {
	opts := &options.Options{
		Top: options.Top{
			Ctx: context.Background(),
		},
	}

	optionsFunc := func(app *cobra.Command) {
		app.SuggestionsMinimumDistance = 1
		app.AddCommand(
			ensure.EnsureCmd(opts),
		)
	}

	return App(version, optionsFunc)
}
