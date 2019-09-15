package demo

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/ensure/demo/petclinic"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Demo(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "demo",
		Short:   "ensuring state of demo applications",
		RunE: func(cmd *cobra.Command, args []string) error {
			return internal.RootAddError
		},
	}

	cmd.AddCommand(
		petclinic.Petclinic(opts, optionsFunc...))
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}
