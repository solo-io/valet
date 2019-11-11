package common

import (
	"os"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"github.com/solo-io/valet/cli/internal/ensure/resource/workflow"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/config"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/options"
)

var (
	MustProvideFileError = errors.Errorf("Must provide file option or subcommand")
)

func LoadConfig(opts *options.Options) (*workflow.Config, error) {
	if opts.Ensure.File == "" {
		return nil, MustProvideFileError
	}

	cfg, err := workflow.LoadConfig(opts.Ensure.Registry, opts.Ensure.File, render.InputParams{})
	if err != nil {
		return nil, err
	}

	if cfg.Cluster != nil {
		if opts.Ensure.GkeClusterName != "" {
			if len(opts.Ensure.GkeClusterName) > 40 {
				opts.Ensure.GkeClusterName = opts.Ensure.GkeClusterName[:40]
			}
			cfg.Cluster.GKE.Name = opts.Ensure.GkeClusterName
		}
	}
	return cfg, nil
}

func LoadEnv(globalConfig *config.ValetGlobalConfig) error {
	for k, v := range globalConfig.Env {
		val := os.Getenv(k)
		if val == "" {
			err := os.Setenv(k, v)
			if err != nil {
				cmd.Stderr().Println("Failed to set environment variable: %s", err.Error())
				return err
			}
		}
	}
	return nil
}

func GetRegistries(globalConfig *config.ValetGlobalConfig) map[string]render.Registry {
	registries := make(map[string]render.Registry)
	for k, v := range globalConfig.Registries {
		registries[k] = v.LocalRegistry
	}
	return registries
}