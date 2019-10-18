package teardown

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/ensure"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Teardown(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "teardown",
		Short: "tears down cluster based on configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return teardown(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.PersistentFlags().StringVarP(&opts.Ensure.File, "file", "f", "", "path to file containing config to ensure")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.GkeClusterName, "gke-cluster-name", "", "", "GKE cluster name to use")
	return cmd
}

func teardown(opts *options.Options) error {
	if opts.Ensure.File == "" {
		return ensure.MustProvideFileError
	}
	cfg, err := resource.LoadConfig(opts.Top.Ctx, opts.Ensure.File)
	if err != nil {
		return err
	}

	if err := ensure.LoadEnv(opts.Top.Ctx); err != nil {
		return err
	}

	if cfg.Cluster != nil {
		if cfg.Cluster.GKE != nil {
			if opts.Ensure.GkeClusterName != "" {
				cfg.Cluster.GKE.Name = opts.Ensure.GkeClusterName
			}
		}
	}

	return TeardownCfg(opts, cfg)
}

func TeardownCfg(opts *options.Options, cfg *resource.Config) error {
	command := cmd.CommandFactory{}
	return cfg.Teardown(opts.Top.Ctx, &command)
}
