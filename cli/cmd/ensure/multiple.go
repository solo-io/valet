package ensure

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Multiple(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "application",
		Short: "ensures application is deployed",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Ensure.File == "" {
				return errors.Errorf("Must provide file to ensure")
			}
			return runMultiple(opts)
		},
	}
	cmd.PersistentFlags().StringToStringVarP(&opts.Ensure.Values, "values", "v", make(map[string]string), "values to provide to application")
	cmd.PersistentFlags().StringSliceVarP(&opts.Ensure.Flags, "flags", "", make([]string, 0), "flags to provide to application")
	cmd.PersistentFlags().BoolVarP(&opts.Ensure.DryRun, "dry-run", "", false, "dry-run and output rendered manifest to stdout")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func runMultiple(opts *options.Options) error {
	input, err := common.LoadInput(opts)
	if err != nil {
		return err
	}

	config, err := common.LoadMultiClusterWorkflow(opts, *input)
	if err != nil {
		return err
	}
	err = config.Ensure(opts.Top.Ctx, *input)
	if opts.Ensure.TeardownOnFinish {
		_ = config.Teardown(opts.Top.Ctx, *input)
	}
	return err
}
