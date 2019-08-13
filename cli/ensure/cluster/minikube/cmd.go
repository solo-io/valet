package minikube

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func MinikubeCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "minikube",
		Short:   "ensures minikube cluster is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			return EnsureMinikube(opts)
		},
	}
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func EnsureMinikube(opts *options.Options) error {
	provisioner := NewMinikubeProvisionerFromOpts(opts.Cluster)
	err := provisioner.Ensure(opts.Top.Ctx)
	if err != nil {
		contextutils.LoggerFrom(opts.Top.Ctx).Errorw("Error ensuring minikube cluster", zap.Error(err))
		return err
	}
	contextutils.LoggerFrom(opts.Top.Ctx).Infow("Minikube is ready")
	return nil
}