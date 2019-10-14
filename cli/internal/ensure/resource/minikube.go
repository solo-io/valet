package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	MinikubeContext = "minikube"
	DefaultKubeVersion = "v1.13.0"
	DefaultCpus = 4
	DefaultMemory = 8192
)

var _ ClusterResource = new(Minikube)

type Minikube struct {}

func (m *Minikube) Ensure(ctx context.Context) error {
	err := cmd.Minikube().Status().Run(ctx)
	if err == nil {
		return m.SetContext(ctx)
	}
	return RunAll(
		ctx,
		cmd.Minikube().Delete().Command(),
		cmd.Minikube().Start().Cpus(DefaultCpus).Memory(DefaultMemory).KubeVersion(DefaultKubeVersion).Command())
}

func (m *Minikube) SetContext(ctx context.Context) error {
	return cmd.Kubectl().UseContext(MinikubeContext).Run(ctx)
}

func (m *Minikube) Teardown(ctx context.Context) error {
	return cmd.Minikube().Delete().Run(ctx)
}



