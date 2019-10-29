package common

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/config"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/options"
	"go.uber.org/zap"
	"os"
)

var (
	MustProvideFileError = errors.Errorf("Must provide file option or subcommand")
)

func LoadApplication(opts *options.Options) (*resource.Application, error) {
	if opts.Ensure.File == "" {
		return nil, MustProvideFileError
	}

	cfg, err := resource.LoadApplication(opts.Top.Ctx, opts.Ensure.File)
	if err != nil {
		return nil, err
	}

	if err := LoadEnv(opts.Top.Ctx); err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadConfig(opts *options.Options) (*resource.Config, error) {
	if opts.Ensure.File == "" {
		return nil, MustProvideFileError
	}

	cfg, err := resource.LoadConfig(opts.Top.Ctx, opts.Ensure.File)
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
		contextutils.LoggerFrom(ctx).Errorw("Failed to load global config", zap.Error(err))
		return err
	}

	for k, v := range globalConfig.Env {
		val := os.Getenv(k)
		if val == "" {
			err := os.Setenv(k, v)
			if err != nil {
				contextutils.LoggerFrom(ctx).Errorw("Failed to set environment variable", zap.Error(err))
				return err
			}
		}
	}
	return nil
}
