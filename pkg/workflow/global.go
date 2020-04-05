package workflow

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"
	"os"
	"path/filepath"
)

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

func GetDefaultGlobalConfigPath() (string, error) {
	valetDir, err := GetValetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(valetDir, "global.yaml"), nil
}

func LoadGlobalConfig(path string, store render.FileStore) (*api.ValetGlobalConfig, error) {
	var c api.ValetGlobalConfig
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &c, nil
	}
	err := store.LoadYaml(path, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func LoadDefaultGlobalConfig(store render.FileStore) (*api.ValetGlobalConfig, error) {
	globalConfigPath, err := GetDefaultGlobalConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadGlobalConfig(globalConfigPath, store)
}
