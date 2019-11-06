package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	MinikubeContext    = "minikube"
	DefaultKubeVersion = "v1.13.0"
	DefaultCpus        = 4
	DefaultMemory      = 8192
)

var _ ClusterResource = new(Minikube)

type Minikube struct{}

func (m *Minikube) Ensure(ctx context.Context, _ InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Ensuring minikube cluster")
	// If minikube status seems healthy, just set context and return
	if err := command.Minikube().Status().SwallowError().Cmd().Run(ctx); err == nil {
		return m.SetContext(ctx, command)
	}
	_ = command.Minikube().Delete().SwallowError().Cmd().Run(ctx)
	return command.Minikube().Start().Cpus(DefaultCpus).Memory(DefaultMemory).KubeVersion(DefaultKubeVersion).Cmd().Run(ctx)
}

func (m *Minikube) SetContext(ctx context.Context, command cmd.Factory) error {
	return command.Kubectl().UseContext(MinikubeContext).Cmd().Run(ctx)
}

func (m *Minikube) Teardown(ctx context.Context, _ InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Tearing down minikube cluster")
	return command.Minikube().Delete().SwallowError().Cmd().Run(ctx)
}
