package gke

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/kube-cluster/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	MissingNameError = errors.Errorf("Must provide a GKE cluster name")
	MissingProjectError = errors.Errorf("Must provide a GKE project")
	MissingLocationError = errors.Errorf("Must provide a GKE location")
)

func GkeCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gke",
		Short:   "ensures GKE cluster is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			return EnsureGke(opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.Cluster.GKE.Name, "name", "n", "", "GKE cluster name")
	cmd.PersistentFlags().StringVarP(&opts.Cluster.GKE.Project, "project", "p", "", "GKE cluster project")
	cmd.PersistentFlags().StringVarP(&opts.Cluster.GKE.Location, "location", "l", "", "GKE cluster location")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func EnsureGke(opts *options.Options) error {
	if err := validateOpts(opts); err != nil {
		return err
	}
	return ensureGke(opts)
}

func ensureGke(opts *options.Options) error {
	provisioner, err := NewGkeProvisionerFromOpts(opts.Top.Ctx, opts.Cluster)
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

func validateOpts(opts *options.Options) error {
	if opts.Cluster.GKE.Name == "" {
		return MissingNameError
	}
	if opts.Cluster.GKE.Project == "" {
		return MissingProjectError
	}
	if opts.Cluster.GKE.Location == "" {
		return MissingLocationError
	}
	return nil
}
