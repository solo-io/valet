package teardown

import (
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/ensure"
	"github.com/solo-io/valet/cli/cmd/ensure/cluster/gke"
	"github.com/solo-io/valet/cli/cmd/ensure/cluster/minikube"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
	return cmd
}

func teardown(opts *options.Options) error {
	cfg, err := ensure.LoadConfig(opts.Top.Ctx, opts.Ensure.File)
	if err != nil {
		return err
	}

	if err := ensure.LoadEnv(opts.Top.Ctx); err != nil {
		return err
	}

	if cfg.Cluster != nil {
		opts.Cluster.Type = cfg.Cluster.Type
		opts.Cluster.GKE = cfg.Cluster.GKE
		opts.Cluster.Minikube = cfg.Cluster.Minikube

		if opts.Cluster.Type == "gke" {
			cluster, err := gke.NewGkeClusterFromOpts(opts.Top.Ctx, opts.Cluster)
			if err != nil {
				return err
			}
			return cluster.Destroy(opts.Top.Ctx)
		} else if opts.Cluster.Type == "minikube" {
			cluster := minikube.NewMinikubeClusterFromOpts(opts.Cluster)
			return cluster.Destroy(opts.Top.Ctx)
		} else {
			return errors.Errorf("unknown type", zap.String("type", opts.Cluster.Type))
		}
	}
	return nil
}
