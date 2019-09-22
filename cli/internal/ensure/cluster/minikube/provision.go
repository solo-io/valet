package minikube

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/internal/ensure/cluster"
	"go.uber.org/zap"
)

var _ cluster.Provisioner = new(minikubeProvisioner)

func NewMinikubeProvisionerFromOpts(cluster *api.Minikube) *minikubeProvisioner {
	minikubeCluster := NewMinikubeClusterFromOpts(cluster)
	return &minikubeProvisioner{
		cluster: minikubeCluster,
	}
}

type minikubeProvisioner struct {
	cluster cluster.KubeCluster
}

func (m *minikubeProvisioner) Ensure(ctx context.Context) error {
	running, err := m.cluster.IsRunning(ctx)
	if err == nil && running {
		contextutils.LoggerFrom(ctx).Infow("Minikube cluster is running")
		return nil
	}
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnw("Error checking if cluster is running, destroying",
			zap.Error(err))
	}
	err = m.cluster.Destroy(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnw("Error destroying cluster",
			zap.Error(err))
		return err
	}
	return m.cluster.Create(ctx)
}

func (m *minikubeProvisioner) GetCluster() cluster.KubeCluster {
	return m.cluster
}
