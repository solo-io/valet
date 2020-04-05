package config

import (
	"fmt"
	"github.com/solo-io/valet/pkg/cmd"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/pkg/cli/options"
	"github.com/spf13/cobra"
)

func Config(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "manage global config for valet",
		RunE: func(cmd *cobra.Command, args []string) error {
			bytes, err := loadGlobalConfigContents(opts)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", string(bytes))
			return nil
		},
	}

	cliutils.ApplyOptions(configCmd, optionsFunc)
	configCmd.AddCommand(SetCmd(opts))
	return configCmd
}

type ValetGlobalConfig struct {
	Env map[string]string `json:"env"`
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
	bytes, err := loadGlobalConfigContents(opts)
	if err != nil {
		return nil, err
	}
	var c ValetGlobalConfig
	if bytes == nil {
		return &c, nil
	}
	if err := yaml.UnmarshalStrict(bytes, &c, yaml.DisallowUnknownFields); err != nil {
		cmd.Stderr().Println("Failed to unmarshal global config: %s", err.Error())
		return nil, err
	}

	return &c, nil
}

func loadGlobalConfigContents(opts *options.Options) ([]byte, error) {
	path, err := GetGlobalConfigPath(opts)
	if err != nil {
		cmd.Stderr().Println("Could not determine global config path")
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cmd.Stdout().Println("No global config exists")
		return nil, nil
	}

	osClient := osutils.NewOsClient()
	bytes, err := osClient.ReadFile(path)
	if err != nil {
		cmd.Stderr().Println("Could not read file %s: %s", path, err.Error())
		return nil, err
	}
	return bytes, nil
}

func StoreGlobalConfig(opts *options.Options, config *ValetGlobalConfig) error {
	path, err := GetGlobalConfigPath(opts)
	if err != nil {
		cmd.Stderr().Println("Could not determine global config path: %s", err.Error())
		return err
	}

	bytes, err := yaml.Marshal(config)
	if err != nil {
		cmd.Stderr().Println("Failed to marshal config: %s", err.Error())
		return err
	}

	if err := ioutil.WriteFile(path, bytes, os.ModePerm); err != nil {
		cmd.Stderr().Println("Failed to write file %s: %s", path, err.Error())
		return err
	}

	return nil
}
