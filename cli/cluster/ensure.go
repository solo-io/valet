package cluster

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/kube-cluster/cli/cluster/gke"
	"github.com/solo-io/kube-cluster/cli/cluster/minikube"
	"github.com/solo-io/kube-cluster/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func EnsureCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ensure",
		Short:   "ensures kubernetes cluster is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RootAddError
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.Cluster.KubeVersion, "kube-version", "v", "v1.13.0", "kube version")
	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.AddCommand(
		GkeCmd(opts),
		MinikubeCmd(opts))
	return cmd
}

func GkeCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gke",
		Short:   "ensures GKE cluster is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ensureGke(opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.Cluster.Name, "name", "n", "", "GKE cluster name")
	cmd.PersistentFlags().StringVarP(&opts.Cluster.Project, "project", "p", "", "GKE cluster project")
	cmd.PersistentFlags().StringVarP(&opts.Cluster.Location, "location", "l", "", "GKE cluster location")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func ensureGke(opts *options.Options) error {
	provisioner, err := gke.NewGkeProvisionerFromOpts(opts.Top.Ctx, opts.Cluster)
	if err != nil {
		contextutils.LoggerFrom(opts.Top.Ctx).Errorw("Error creating gke provisioner", zap.Error(err))
		return err
	}
	err = provisioner.Ensure(opts.Top.Ctx)
	if err != nil {
		contextutils.LoggerFrom(opts.Top.Ctx).Errorw("Error ensuring gke cluster", zap.Error(err))
		return err
	}
	contextutils.LoggerFrom(opts.Top.Ctx).Infow("gke is ready")
	return nil
}

func MinikubeCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "minikube",
		Short:   "ensures minikube cluster is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ensureMinikube(opts)
		},
	}
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func ensureMinikube(opts *options.Options) error {
	provisioner := minikube.NewMinikubeProvisionerFromOpts(opts.Top.Ctx, opts.Cluster)
	err := provisioner.Ensure(opts.Top.Ctx)
	if err != nil {
		contextutils.LoggerFrom(opts.Top.Ctx).Errorw("Error ensuring minikube cluster", zap.Error(err))
		return err
	}
	contextutils.LoggerFrom(opts.Top.Ctx).Infow("Minikube is ready")
	return nil
}