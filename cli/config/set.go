package config

import (
	"context"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"strings"
)

func SetCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "set one or more config values (foo=bar)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setConfig(opts.Top.Ctx, args)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func setConfig(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.Errorf("Must provide at least one argument, i.e. 'foo=bar' (without single quotes)")
	}

	config, err := LoadGlobalConfig(ctx)
	if err != nil {
		return err
	}

	if config.Env == nil {
		config.Env = make(map[string]string)
	}

	for _, arg := range args {
		splitArg := strings.Split(arg, "=")
		if len(splitArg) != 2 {
			return errors.Errorf("Args must be of the form 'foo=bar' (without single quotes)")
		}
		config.Env[splitArg[0]] = splitArg[1]
	}

	err = StoreGlobalConfig(ctx, config)
	if err != nil {
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("Successfully updated config")
	return nil
}
