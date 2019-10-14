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
	// If minikube status seems healthy, just set context and return
	if err := cmd.Minikube().Status().SwallowError().Run(ctx); err == nil {
		return m.SetContext(ctx)
	}
	_ = cmd.Minikube().Delete().SwallowError().Run(ctx)
	return cmd.Minikube().Start().Cpus(DefaultCpus).Memory(DefaultMemory).KubeVersion(DefaultKubeVersion).Run(ctx)
}

func (m *Minikube) SetContext(ctx context.Context) error {
	return cmd.Kubectl().UseContext(MinikubeContext).Run(ctx)
}

func (m *Minikube) Teardown(ctx context.Context) error {
	return cmd.Minikube().Delete().SwallowError().Run(ctx)
}



