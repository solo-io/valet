package ensure

import (
	"context"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/cmd/config"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

var (
	MustProvideFileError = errors.Errorf("Must provide file option or subcommand")
)

func Ensure(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ensure",
		Short: "ensures kubernetes cluster is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ensure(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.PersistentFlags().StringVarP(&opts.Ensure.File, "file", "f", "", "path to file containing config to ensure")
	cmd.PersistentFlags().BoolVarP(&opts.Ensure.ValetArtifacts, "valet-artifacts", "", false, "use valet artifacts (in google storage)")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.GlooVersion, "gloo-version", "", "", "gloo version")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.GkeClusterName, "gke-cluster-name", "", "", "GKE cluster name to use")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.LocalArtifactsDir, "local-artifacts-dir", "", "", "local directory containing artifacts")
	return cmd
}

func ensure(opts *options.Options) error {
	if opts.Ensure.File == "" {
		return MustProvideFileError
	}

	cfg, err := api.LoadConfig(opts.Top.Ctx, opts.Ensure.File)
	if err != nil {
		return err
	}

	if err := LoadEnv(opts.Top.Ctx); err != nil {
		return err
	}

	if cfg.Cluster != nil {
		if opts.Ensure.GkeClusterName != "" {
			cfg.Cluster.GKE.Name = opts.Ensure.GkeClusterName
		}
	}
	if cfg.Gloo != nil {
		if opts.Ensure.GlooVersion != "" {
			cfg.Gloo.Version = opts.Ensure.GlooVersion
		}
		cfg.Gloo.ValetArtifacts = opts.Ensure.ValetArtifacts
		cfg.Gloo.LocalArtifactsDir = opts.Ensure.LocalArtifactsDir
	}

	return cfg.Ensure(opts.Top.Ctx)
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
