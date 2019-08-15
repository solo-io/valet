package config

import (
	"context"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func SetCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "set a config value",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setConfig(opts.Top.Ctx, args)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func setConfig(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return errors.Errorf("Must provide exactly two arguments, key and value")
	}

	config, err := LoadGlobalConfig(ctx)
	if err != nil {
		return err
	}

	if config.Env == nil {
		config.Env = make(map[string]string)
	}
	config.Env[args[0]] = args[1]

	err = StoreGlobalConfig(ctx, config)
	if err != nil {
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("Successfully updated config")
	return nil
}
