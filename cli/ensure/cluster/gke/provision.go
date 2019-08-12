package gke

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/kube-cluster/cli/ensure/cluster/cluster"
	"github.com/solo-io/kube-cluster/cli/options"
	"go.uber.org/zap"
)

var _ cluster.Provisioner = new(gkeProvisioner)

func NewGkeProvisionerFromOpts(ctx context.Context, opts options.Cluster) (*gkeProvisioner, error) {
	gkeCluster, err := NewGkeClusterFromOpts(ctx, opts)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error getting gke cluster", zap.Error(err))
		return nil, err
	}
	return &gkeProvisioner{
		cluster: gkeCluster,
	}, nil
}

type gkeProvisioner struct {
	cluster cluster.KubeCluster
}

func (m *gkeProvisioner) Ensure(ctx context.Context) error {
	running, err := m.cluster.IsRunning(ctx)
	if err == nil && running {
		contextutils.LoggerFrom(ctx).Infow("GKE cluster is running")
		return nil
	} else if err != nil {
		contextutils.LoggerFrom(ctx).Warnw("Error checking if cluster is running, cleaning up any existing cluster")
	} else {
		contextutils.LoggerFrom(ctx).Infow("GKE cluster not running")
	}
	err = m.cluster.Destroy(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnw("Error destroying cluster",
			zap.Error(err))
	}
	return m.cluster.Create(ctx)
}

func (m *gkeProvisioner) GetCluster() cluster.KubeCluster {
	return m.cluster
}
