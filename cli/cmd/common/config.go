package common

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/config"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/options"
)

var (
	MustProvideFileError = errors.Errorf("Must provide file option or subcommand")
)

func LoadApplication(opts *options.Options, input resource.InputParams) (*resource.Application, error) {
	if opts.Ensure.File == "" {
		return nil, MustProvideFileError
	}

	ref := resource.ApplicationRef{
		Path:   opts.Ensure.File,
	}

	app, err := ref.Load(opts.Top.Ctx, input)
	if err != nil {
		return nil, err
	}

	if err := LoadEnv(opts.Top.Ctx); err != nil {
		return nil, err
	}
	return app, nil
}

func LoadConfig(opts *options.Options) (*resource.Config, error) {
	if opts.Ensure.File == "" {
		return nil, MustProvideFileError
	}

	cfg, err := resource.LoadConfig(opts.Ensure.File)
	if err != nil {
		return nil, err
	}

	if err := LoadEnv(opts.Top.Ctx); err != nil {
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

func LoadEnv(ctx context.Context) error {
	globalConfig, err := config.LoadGlobalConfig(ctx)
	if err != nil {
		cmd.Stderr().Println("Failed to load global config: %s", err.Error())
		return err
	}

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
