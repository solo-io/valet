package config

import (
	"strings"

	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func SetCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "set one or more config values (foo=bar)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setConfig(opts, args)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func setConfig(opts *options.Options, args []string) error {
	if len(args) == 0 {
		return errors.Errorf("Must provide at least one argument, i.e. 'foo=bar' (without single quotes)")
	}

	config, err := LoadGlobalConfig(opts)
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

	err = StoreGlobalConfig(opts, config)
	if err != nil {
		return err
	}
	cmd.Stdout(opts.Top.Ctx).Printf("Successfully updated config")
	return nil
}
