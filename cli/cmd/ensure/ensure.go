package ensure

import (
	"context"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/config"
	"github.com/solo-io/valet/cli/cmd/teardown"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
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
			err := ensure(opts)
			if opts.Ensure.TeardownOnFinish {
				cfg, teardownErr := LoadConfig(opts)
				if teardownErr != nil {
					contextutils.LoggerFrom(opts.Top.Ctx).Errorw("Error trying to teardown")
				} else {
					_ = teardown.TeardownCfg(opts, cfg)
				}
			}
			return err
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.PersistentFlags().StringVarP(&opts.Ensure.File, "file", "f", "", "path to file containing config to ensure")
	cmd.PersistentFlags().BoolVarP(&opts.Ensure.ValetArtifacts, "valet-artifacts", "", false, "use valet artifacts (in google storage)")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.GlooVersion, "gloo-version", "", "", "gloo version")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.GkeClusterName, "gke-cluster-name", "", "", "GKE cluster name to use")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.LocalArtifactsDir, "local-artifacts-dir", "", "", "local directory containing artifacts")
	cmd.PersistentFlags().BoolVarP(&opts.Ensure.TeardownOnFinish, "teardown-on-finish", "", false, "attempt teardown before exit. return code should be 0 if ensure succeeded, nonzero otherwise")
	return cmd
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
	if cfg.Gloo != nil {
		if opts.Ensure.GlooVersion != "" {
			cfg.Gloo.Version = opts.Ensure.GlooVersion
		}
		cfg.Gloo.ValetArtifacts = opts.Ensure.ValetArtifacts
		cfg.Gloo.LocalArtifactsDir = opts.Ensure.LocalArtifactsDir
	}
	return cfg, nil
}

func ensure(opts *options.Options) error {
	cfg, err := LoadConfig(opts)
	if err != nil {
		return err
	}
	command := cmd.CommandFactory{}
	return cfg.Ensure(opts.Top.Ctx, &command)
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
