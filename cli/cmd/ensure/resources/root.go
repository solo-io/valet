package resources

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Resources(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resources",
		Short:   "ensuring resources are applied to the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return EnsureResources(opts)
		},
	}

	cmd.PersistentFlags().StringSliceVarP(&opts.Ensure.Resources, "resource", "r", nil, "one or more resources to pass to 'kubectl apply'")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func EnsureResources(opts *options.Options) error {
	for _, resource := range opts.Ensure.Resources {
		out, err := internal.ExecuteCmd("kubectl", "apply", "-f", resource)
		if err != nil {
			contextutils.LoggerFrom(opts.Top.Ctx).Errorw("Error applying resources to cluster",
				zap.Error(err),
				zap.String("out", out),
				zap.String("resource", resource))
			return err
		}
	}
	return nil
}
