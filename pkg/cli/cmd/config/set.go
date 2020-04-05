package config

import (
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/workflow"
	"strings"

	errors "github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/pkg/cli/options"
	"github.com/spf13/cobra"
)

func SetCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "set one or more config values (foo=bar)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setConfig(opts, args)
		},
	}

	cliutils.ApplyOptions(setCmd, optionsFunc)
	return setCmd
}

func setConfig(opts *options.Options, args []string) error {
	if len(args) == 0 {
		return errors.Errorf("Must provide at least one argument, i.e. 'foo=bar' (without single quotes)")
	}

	globalConfigPath := opts.Config.GlobalConfigPath
	if globalConfigPath == "" {
		defaultPath, err := workflow.GetDefaultGlobalConfigPath()
		if err != nil {
			return err
		}
		globalConfigPath = defaultPath
	}
	fileStore := render.NewFileStore()

	config, err := workflow.LoadGlobalConfig(globalConfigPath, fileStore)
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

	err = fileStore.SaveYaml(globalConfigPath, config)
	if err != nil {
		return err
	}
	cmd.Stdout().Println("Successfully updated config")
	return nil
}
