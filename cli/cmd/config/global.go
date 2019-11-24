package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func Config(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	configCommand := &cobra.Command{
		Use:   "config",
		Short: "manage global config for valet",
		RunE: func(cmd *cobra.Command, args []string) error {
			return internal.RootAddError
		},
	}

	cliutils.ApplyOptions(configCommand, optionsFunc)
	configCommand.AddCommand(SetCmd(opts))
	configCommand.AddCommand(RegistryCmd(opts))
	return configCommand
}

type ValetGlobalConfig struct {
	Env        map[string]string        `yaml:"env"`
	Registries map[string]ValetRegistry `yaml:"registries"`
}

type ValetRegistry struct {
	DirectoryRegistry *render.DirectoryRegistry `yaml:"directory"`
}

func (v *ValetRegistry) GetType() string {
	if v.DirectoryRegistry != nil {
		return "Directory"
	}
	return "Unknown"
}

func GetValetConfigDir() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	valetDir := filepath.Join(userHome, ".valet")
	if _, err := os.Stat(valetDir); os.IsNotExist(err) {
		err = os.Mkdir(valetDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return valetDir, nil
}

func GetGlobalConfigPath(opts *options.Options) (string, error) {
	if opts.Config.GlobalConfigPath != "" {
		return opts.Config.GlobalConfigPath, nil
	}
	valetDir, err := GetValetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(valetDir, "global.yaml"), nil
}

func LoadGlobalConfig(opts *options.Options) (*ValetGlobalConfig, error) {
	var c ValetGlobalConfig
	path, err := GetGlobalConfigPath(opts)
	if err != nil {
		cmd.Stderr(context.TODO()).Printf("Could not determine global config path")
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cmd.Stdout(context.TODO()).Printf("No global config exists")
		return &c, nil
	}

	osClient := osutils.NewOsClient()
	bytes, err := osClient.ReadFile(path)
	if err != nil {
		cmd.Stderr(context.TODO()).Printf("Could not read file %s: %s", path, err.Error())
		return nil, err
	}

	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		cmd.Stderr(context.TODO()).Printf("Failed to unmarshal file %s: %s", path, err.Error())
		return nil, err
	}

	return &c, nil
}

func StoreGlobalConfig(opts *options.Options, config *ValetGlobalConfig) error {
	path, err := GetGlobalConfigPath(opts)
	if err != nil {
		cmd.Stderr(context.TODO()).Printf("Could not determine global config path: %s", err.Error())
		return err
	}

	bytes, err := yaml.Marshal(config)
	if err != nil {
		cmd.Stderr(context.TODO()).Printf("Failed to marshal config: %s", err.Error())
		return err
	}

	if err := ioutil.WriteFile(path, bytes, os.ModePerm); err != nil {
		cmd.Stderr(context.TODO()).Printf("Failed to write file %s: %s", path, err.Error())
		return err
	}

	return nil
}
