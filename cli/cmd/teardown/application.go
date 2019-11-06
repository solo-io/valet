package teardown

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Application(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "teardown",
		Short: "tears down application based on configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Ensure.File == "" {
				return errors.Errorf("Must provide file to ensure")
			}
			return TeardownApplication(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func TeardownApplication(opts *options.Options) error {
	input := resource.InputParams{
		Values: opts.Ensure.Values,
		Flags:  opts.Ensure.Flags,
	}
	cfg, err := common.LoadApplication(opts, input)
	if err != nil {
		return err
	}
	command := cmd.CommandFactory{}
	return cfg.Teardown(opts.Top.Ctx, input, &command)
}
