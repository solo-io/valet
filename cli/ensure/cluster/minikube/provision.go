package minikube

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/kube-cluster/cli/ensure/cluster/cluster"
	"github.com/solo-io/kube-cluster/cli/options"
	"go.uber.org/zap"
)

var _ cluster.Provisioner = new(minikubeProvisioner)

func NewMinikubeProvisionerFromEnv(ctx context.Context) (*minikubeProvisioner, error) {
	minikubeCluster, err := NewMinikubeClusterFromEnv(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error getting minikube cluster", zap.Error(err))
		return nil, err
	}
	return &minikubeProvisioner{
		cluster: minikubeCluster,
	}, nil
}

func NewMinikubeProvisionerFromOpts(ctx context.Context, opts options.Cluster) *minikubeProvisioner {
	minikubeCluster := NewMinikubeClusterFromOpts(ctx, opts)
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





