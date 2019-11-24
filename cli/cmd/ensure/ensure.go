package ensure

import (
	"context"

	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/teardown"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
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
				if teardownErr := teardown.TeardownCfg(opts); teardownErr != nil {
					cmd.Stderr(context.TODO()).Println("error trying to teardown: %s", teardownErr.Error())
				}
			}
			return err
		},
	}

	cliutils.ApplyOptions(ensureCmd, optionsFunc)
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.Registry, "registry", "r", "default", "registry name")
	ensureCmd.PersistentFlags().BoolVarP(&opts.Ensure.Step, "step", "s", false, "wait for user input between steps or resources")
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.File, "file", "f", "", "path to file containing config to ensure")
	ensureCmd.PersistentFlags().BoolVarP(&opts.Ensure.ValetArtifacts, "valet-artifacts", "", false, "use valet artifacts (in google storage)")
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.GlooVersion, "gloo-version", "", "", "gloo version")
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.KubeConfig, "kubeconfig", "", "$HOME/.kube/config", "kubeconfig")
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.GkeClusterName, "gke-cluster-name", "", "", "GKE cluster name to use")
	ensureCmd.PersistentFlags().StringVarP(&opts.Ensure.LocalArtifactsDir, "local-artifacts-dir", "", "", "local directory containing artifacts")
	ensureCmd.PersistentFlags().BoolVarP(&opts.Ensure.TeardownOnFinish, "teardown-on-finish", "", false, "attempt teardown before exit. return code should be 0 if ensure succeeded, nonzero otherwise")

	ensureCmd.AddCommand(Application(opts))
	ensureCmd.AddCommand(Multiple(opts))

	return ensureCmd
}

func ensure(opts *options.Options) error {
	input, err := common.LoadInput(opts)
	if err != nil {
		return err
	}
	cfg, err := common.LoadClusterWorkflow(opts, *input)
	if err != nil {
		return err
	}
	return cfg.Ensure(opts.Top.Ctx, *input)
}
