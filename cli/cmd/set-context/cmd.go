package set_context

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

var (
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
	input, err := common.LoadInput(opts)
	if err != nil {
		return err
	}
	cfg, err := common.LoadClusterWorkflow(opts, *input)
	if err != nil {
		return err
	}
	if cfg.Cluster == nil {
		return MustSpecifyClusterError
	}
	if err := input.RenderFields(cfg.Cluster); err != nil {
		return err
	}
	// TODO(EItanya): clean up deref
	return cfg.Cluster.SetContext(opts.Top.Ctx, *input)
}
