package set_context

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/ensure"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

var (
	MustProvideFileError    = errors.Errorf("Must provide file option")
	MustSpecifyClusterError = errors.Errorf("Must specify cluster")
)

func SetContext(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-context",
		Short: "sets context based on a provided config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setContext(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.PersistentFlags().StringVarP(&opts.Ensure.File, "file", "f", "", "path to file containing config to ensure")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.GkeClusterName, "gke-cluster-name", "", "", "GKE cluster name to use")
	return cmd
}

func setContext(opts *options.Options) error {
	if opts.Ensure.File == "" {
		return MustProvideFileError
	}
	cfg, err := resource.LoadConfig(opts.Top.Ctx, opts.Ensure.File)
	if err != nil {
		return err
	}
	if err := ensure.LoadEnv(opts.Top.Ctx); err != nil {
		return err
	}
	if cfg.Cluster == nil {
		return MustSpecifyClusterError
	}

	if cfg.Cluster.GKE != nil {
		if opts.Ensure.GkeClusterName != "" {
			cfg.Cluster.GKE.Name = opts.Ensure.GkeClusterName
		}
		if cfg.Cluster.GKE.Name == "" {
			return MustSpecifyClusterError
		}
	}
	command := cmd.CommandFactory{}
	return cfg.Cluster.SetContext(opts.Top.Ctx, &command)
}
