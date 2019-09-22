package minikube

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cluster"
	"go.uber.org/zap"
	"strings"
)

var _ cluster.KubeCluster = new(minikubeCluster)

func NewMinikubeClusterFromOpts(cluster *api.Minikube) *minikubeCluster {
	return &minikubeCluster{
		cluster: cluster,
	}
}

type minikubeCluster struct {
	cluster *api.Minikube
}

func (m *minikubeCluster) KubeVersion(ctx context.Context) string {
	return m.cluster.KubeVersion
}

func (m *minikubeCluster) IsRunning(ctx context.Context) (bool, error) {
	err := m.SetKubeContext(ctx)
	if err != nil {
		return false, err
	}
	out, err := internal.ExecuteCmd("minikube", "status")
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnw("Error checking minikube status", zap.String("out", out))
		return false, err
	}
	return true, nil
}

func (m *minikubeCluster) SetKubeContext(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Setting kube context to minikube")
	out, err := internal.ExecuteCmd("kubectl", "config", "use-context", "minikube")
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error setting kube context to minikube",
			zap.Error(err),
			zap.String("output", out))
	}
	return err
}

func (m *minikubeCluster) Create(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Creating minikube")
	out, err := internal.ExecuteCmd("minikube", "start", "--memory=8192", "--cpus=4", "--kubernetes-version="+m.cluster.KubeVersion)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating minikube",
			zap.Error(err),
			zap.String("output", out))
		return err
	}
	return m.SetKubeContext(ctx)
}

func (m *minikubeCluster) Destroy(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Destroying minikube")
	out, err := internal.ExecuteCmd("minikube", "delete")
	if err != nil {
		if !strings.Contains(out, "Docker machine \"minikube\" does not exist.") {
			contextutils.LoggerFrom(ctx).Errorw("Error destroying minikube",
				zap.Error(err),
				zap.String("output", out))
			return err
		}
	}
	return nil
}
