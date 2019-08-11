package minikube

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/kube-cluster/cli/cluster/cluster"
	"github.com/solo-io/kube-cluster/cli/internal"
	"github.com/solo-io/kube-cluster/cli/options"
	"go.uber.org/zap"
	"strings"
)

var _ cluster.KubeCluster = new(minikubeCluster)

type MinikubeClusterConfig struct {
	KubeVersion     string `split_words:"true" default:"v1.13.0"`
}

func NewMinikubeClusterFromEnv(ctx context.Context) (*minikubeCluster, error) {
	var config MinikubeClusterConfig
	err := envconfig.Process("", &config)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error parsing env config", zap.Error(err))
		return nil, err
	}
	return &minikubeCluster{
		config: config,
	}, nil
}

func NewMinikubeClusterFromOpts(ctx context.Context, opts options.Cluster) *minikubeCluster {
	config := MinikubeClusterConfig{
		KubeVersion: opts.KubeVersion,
	}
	return &minikubeCluster{
		config: config,
	}
}

type minikubeCluster struct {
	config MinikubeClusterConfig
}

func (m *minikubeCluster) KubeVersion(ctx context.Context) string {
	return m.config.KubeVersion
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
	out, err := internal.ExecuteCmd("minikube", "update-context")
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error setting kube context to minikube",
			zap.Error(err),
			zap.String("output", out))
	}
	return err
}

func (m *minikubeCluster) Create(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Creating minikube")
	out, err := internal.ExecuteCmd("minikube", "start", "--memory=8192", "--cpus=4", "--kubernetes-version=" + m.config.KubeVersion)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating minikube",
			zap.Error(err),
			zap.String("output", out))
	}
	return err
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


