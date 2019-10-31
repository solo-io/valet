package ensure

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/teardown"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
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

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func ensureApplication(opts *options.Options) error {
	cfg, err := common.LoadApplication(opts)
	if err != nil {
		return err
	}
	command := cmd.CommandFactory{}
	return cfg.Ensure(opts.Top.Ctx, &command)
}
