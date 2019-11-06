package ensure

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/teardown"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Ensure(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	ensureCmd := &cobra.Command{
		Use:   "ensure",
		Short: "ensures kubernetes cluster is running",
		RunE: func(_ *cobra.Command, args []string) error {
			err := ensure(opts)
			if opts.Ensure.TeardownOnFinish {
				cfg, teardownErr := common.LoadConfig(opts)
				if teardownErr != nil {
					cmd.Stderr().Println("error trying to teardown: %s", teardownErr.Error())
				} else {
					_ = teardown.TeardownCfg(opts, cfg)
				}
			}
			return err
		},
	}

	cliutils.ApplyOptions(ensureCmd, optionsFunc)
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.File, "file", "f", "", "path to file containing config to ensure")
	ensureCmd.PersistentFlags().BoolVarP(&opts.Ensure.ValetArtifacts, "valet-artifacts", "", false, "use valet artifacts (in google storage)")
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.GlooVersion, "gloo-version", "", "", "gloo version")
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.GkeClusterName, "gke-cluster-name", "", "", "GKE cluster name to use")
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.LocalArtifactsDir, "local-artifacts-dir", "", "", "local directory containing artifacts")
	ensureCmd.PersistentFlags().BoolVarP(&opts.Ensure.TeardownOnFinish, "teardown-on-finish", "", false, "attempt teardown before exit. return code should be 0 if ensure succeeded, nonzero otherwise")

	ensureCmd.AddCommand(Application(opts))

	return ensureCmd
}

func ensure(opts *options.Options) error {
	input := resource.InputParams{
		Values: opts.Ensure.Values,
		Flags: opts.Ensure.Flags,
	}
	cfg, err := common.LoadConfig(opts)
	if err != nil {
		return err
	}
	command := cmd.CommandFactory{}
	return cfg.Ensure(opts.Top.Ctx, input, &command)
}
