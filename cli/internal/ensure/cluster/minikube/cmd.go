package minikube

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/api"
	"go.uber.org/zap"
)

func EnsureMinikube(ctx context.Context, cluster *api.Minikube) error {
	provisioner := NewMinikubeProvisionerFromOpts(cluster)
	err := provisioner.Ensure(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error ensuring minikube cluster", zap.Error(err))
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("Minikube is ready")
	return nil
}
