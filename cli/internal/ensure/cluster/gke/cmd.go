package gke

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/api"
	"go.uber.org/zap"
)

var (
	MissingNameError = errors.Errorf("Must provide a GKE cluster name")
	MissingProjectError = errors.Errorf("Must provide a GKE project")
	MissingLocationError = errors.Errorf("Must provide a GKE location")
)

func EnsureGke(ctx context.Context, cluster *api.GKE) error {
	if err := validateOpts(cluster); err != nil {
		return err
	}
	return ensureGke(ctx, cluster)
}

func ensureGke(ctx context.Context, cluster *api.GKE) error {
	provisioner, err := NewGkeProvisionerFromOpts(ctx, cluster)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating gke provisioner", zap.Error(err))
		return err
	}
	err = provisioner.Ensure(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error ensuring gke cluster", zap.Error(err))
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("gke is ready")
	return nil
}

func validateOpts(cluster *api.GKE) error {
	if cluster.Name == "" {
		return MissingNameError
	}
	if cluster.Project == "" {
		return MissingProjectError
	}
	if cluster.Location == "" {
		return MissingLocationError
	}
	return nil
}
