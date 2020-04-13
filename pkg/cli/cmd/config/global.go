package config

import (
	"fmt"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/pkg/cli/options"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/workflow"
	"github.com/spf13/cobra"
)

func Config(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "manage global config for valet",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalConfigPath := opts.Config.GlobalConfigPath
			if globalConfigPath == "" {
				defaultPath, err := workflow.GetDefaultGlobalConfigPath()
				if err != nil {
					return err
				}
				globalConfigPath = defaultPath
			}
			fileStore := render.NewFileStore()
			if exists, err := fileStore.Exists(globalConfigPath); err != nil || !exists {
				return err
			}
			contents, err := fileStore.Load(globalConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", contents)
			return nil
		},
	}

	cliutils.ApplyOptions(configCmd, optionsFunc)
	configCmd.AddCommand(SetCmd(opts))
	return configCmd
}
