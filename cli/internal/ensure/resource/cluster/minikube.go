package cluster

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

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

func (m *Minikube) Ensure(ctx context.Context, input render.InputParams) error {
	cmd.Stdout().Println("Ensuring minikube cluster")
	// If minikube status seems healthy, just set context and return
	if err := input.Runner().Run(ctx, cmd.New().Minikube().Status().SwallowError().Cmd()); err == nil {
		return m.SetContext(ctx, input.Runner())
	}
	_ = input.Runner().Run(ctx, cmd.New().Minikube().Delete().SwallowError().Cmd())
	return input.Runner().Run(ctx, cmd.New().Minikube().Start().Cpus(DefaultCpus).Memory(DefaultMemory).KubeVersion(DefaultKubeVersion).Cmd())
}

func (m *Minikube) SetContext(ctx context.Context, runner cmd.Runner) error {
	return runner.Run(ctx, cmd.New().Kubectl().UseContext(MinikubeContext).Cmd())
}

func (m *Minikube) Teardown(ctx context.Context, input render.InputParams) error {
	cmd.Stdout().Println("Tearing down minikube cluster")
	return input.Runner().Run(ctx, cmd.New().Minikube().Delete().SwallowError().Cmd())
}
