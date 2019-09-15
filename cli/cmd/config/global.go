package config

import (
	"context"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Config(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "global config for valet",
		RunE: func(cmd *cobra.Command, args []string) error {
			return internal.RootAddError
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.AddCommand(SetCmd(opts))
	return cmd
}

type ValetGlobalConfig struct {
	Env map[string]string `yaml:"env"`
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

func GetGlobalConfigPath() (string, error) {
	valetDir, err := GetValetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(valetDir, "global.yaml"), nil
}

func LoadGlobalConfig(ctx context.Context) (*ValetGlobalConfig, error) {
	var c ValetGlobalConfig
	path, err := GetGlobalConfigPath()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not determine global config path", zap.Error(err))
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		contextutils.LoggerFrom(ctx).Infow("No global config exists")
		return &c, nil
	}

	osClient := osutils.NewOsClient()
	bytes, err := osClient.ReadFile(path)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not read file",
			zap.Error(err),
			zap.String("path", path))
		return nil, err
	}

	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to unmarshal file",
			zap.Error(err),
			zap.String("path", path),
			zap.ByteString("bytes", bytes))
		return nil, err
	}

	return &c, nil
}

func StoreGlobalConfig(ctx context.Context, config *ValetGlobalConfig) error {
	path, err := GetGlobalConfigPath()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not determine global config path", zap.Error(err))
		return err
	}

	bytes, err := yaml.Marshal(config)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to marshal config", zap.Error(err), zap.Any("config", config))
		return err
	}

	if err := ioutil.WriteFile(path, bytes, os.ModePerm); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to write file",
			zap.Error(err),
			zap.String("path", path),
			zap.ByteString("bytes", bytes))
		return err
	}

	return nil
}
