package ensure

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/teardown"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Ensure(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ensure",
		Short: "ensures kubernetes cluster is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ensure(opts)
			if opts.Ensure.TeardownOnFinish {
				cfg, teardownErr := common.LoadConfig(opts)
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

	cmd.AddCommand(Application(opts))

	return cmd
}

func ensure(opts *options.Options) error {
	cfg, err := common.LoadConfig(opts)
	if err != nil {
		return err
	}
	command := cmd.CommandFactory{}
	return cfg.Ensure(opts.Top.Ctx, &command)
}
