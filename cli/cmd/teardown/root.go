package teardown

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Teardown(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "teardown",
		Short: "tears down cluster based on configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return TeardownCfg(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.PersistentFlags().StringVarP(&opts.Ensure.Registry, "registry", "r", "default", "registry name")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.File, "file", "f", "", "path to file containing config to ensure")
	cmd.PersistentFlags().StringVarP(&opts.Ensure.GkeClusterName, "gke-cluster-name", "", "", "GKE cluster name to use")

	cmd.AddCommand(Application(opts))
	return cmd
}

func TeardownCfg(opts *options.Options) error {
	input, err := common.LoadInput(opts)
	if err != nil {
		return err
	}
	cfg, err := common.LoadClusterWorkflow(opts, *input)
	if err != nil {
		return err
	}
	return cfg.Teardown(opts.Top.Ctx, *input)
}
