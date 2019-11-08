package teardown

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Application(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "teardown",
		Short: "tears down application based on configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return TeardownApplication(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func TeardownApplication(opts *options.Options) error {
	input := render.InputParams{
		Values: opts.Ensure.Values,
		Flags:  opts.Ensure.Flags,
	}
	if opts.Ensure.File == "" {
		return common.MustProvideFileError
	}
	ref := application.Ref{
		Path: opts.Ensure.File,
	}
	command := cmd.CommandFactory{}
	return ref.Teardown(opts.Top.Ctx, input, &command)
}
