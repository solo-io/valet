package ensure

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/teardown"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Application(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "application",
		Short: "ensures application is deployed",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Ensure.File == "" {
				return errors.Errorf("Must provide file to ensure")
			}
			err := ensureApplication(opts)
			if opts.Ensure.TeardownOnFinish {
				_ = teardown.TeardownApplication(opts)
			}
			return err
		},
	}
	cmd.PersistentFlags().StringToStringVarP(&opts.Ensure.Values, "values", "v", make(map[string]string), "values to provide to application")
	cmd.PersistentFlags().StringSliceVarP(&opts.Ensure.Flags, "flags", "", make([]string, 0), "flags to provide to application")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func ensureApplication(opts *options.Options) error {
	input := resource.InputParams{
		Values: opts.Ensure.Values,
		Flags:  opts.Ensure.Flags,
		Step:   opts.Ensure.Step,
	}
	if opts.Ensure.File == "" {
		return common.MustProvideFileError
	}
	ref := resource.ApplicationRef{
		Path: opts.Ensure.File,
	}
	command := cmd.CommandFactory{}
	return ref.Ensure(opts.Top.Ctx, input, &command)
}
