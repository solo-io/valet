package config

import (
	"fmt"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func RegistryCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "manage valet registries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return internal.RootAddError
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.Config.RegistryName, "registry-name", "", "", "name of registry")
	cmd.PersistentFlags().StringVarP(&opts.Config.RegistryPath, "registry-path", "", "", "path of registry")
	cmd.AddCommand(ListRegistriesCmd(opts))
	cmd.AddCommand(AddRegistryCmd(opts))
	cmd.AddCommand(DeleteRegistryCmd(opts))
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func ListRegistriesCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list current valet registries",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalConfig, err := LoadGlobalConfig(opts)
			if err != nil {
				return err
			}
			if len(globalConfig.Registries) == 0 {
				fmt.Printf("Valet currently has no registries configured. Use `valet registry add` to add one.\n")
				return nil
			}
			fmt.Println("Current valet registries:")
			for k, v := range globalConfig.Registries {
				fmt.Printf("%s: %s -> %s\n", v.GetType(), k, v.DirectoryRegistry.Path)
			}
			return nil
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func AddRegistryCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add registry to valet",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalConfig, err := LoadGlobalConfig(opts)
			if err != nil {
				return err
			}
			if globalConfig.Registries == nil {
				globalConfig.Registries = make(map[string]ValetRegistry)
			}
			if opts.Config.RegistryName == "" {
				return errors.Errorf("Must provide registry-name")
			}
			if opts.Config.RegistryPath == "" {
				return errors.Errorf("Must provide registry-path")
			}
			globalConfig.Registries[opts.Config.RegistryName] = ValetRegistry{
				DirectoryRegistry: &render.DirectoryRegistry{
					Path: opts.Config.RegistryPath,
				},
			}
			err = StoreGlobalConfig(opts, globalConfig)
			if err != nil {
				return err
			}
			fmt.Println("Successfully updated config")
			return nil
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}
func DeleteRegistryCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete registry from valet",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Config.RegistryName == "" {
				return errors.Errorf("Must provide registry-name")
			}
			globalConfig, err := LoadGlobalConfig(opts)
			if err != nil {
				return err
			}
			if globalConfig.Registries == nil {
				return errors.Errorf("Registry not found")
			}
			if _, ok := globalConfig.Registries[opts.Config.RegistryName]; !ok {
				return errors.Errorf("Registry not found")
			} else {
				delete(globalConfig.Registries, opts.Config.RegistryName)
			}
			err = StoreGlobalConfig(opts, globalConfig)
			if err != nil {
				return err
			}
			fmt.Println("Successfully updated config")
			return nil
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}
