package cluster

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

const (
	MinikubeContext    = "minikube"
)

var _ ClusterResource = new(Minikube)

type Minikube struct {
	Cpus        int    `yaml:"cpus" valet:"template,key=MinikubeCpus,default=4"`
	Memory      int    `yaml:"memory" valet:"template,key=MinikubeMemory,default=8192"`
	KubeVersion string `yaml:"name" valet:"template,key=KubeVersion,default=v1.13.0"`
	VmDriver    string `yaml:"vmDriver" valet:"template,key=MinikubeVmDriver,default=virtualbox"`
}

func (m *Minikube) Ensure(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(m); err != nil {
		return err
	}
	cmd.Stdout(ctx).Printf("Ensuring minikube cluster (cpus: %d, memory: %d, version: %s, driver: %s)",
		m.Cpus, m.Memory, m.KubeVersion, m.VmDriver)
	// If minikube status seems healthy, just set context and return
	if err := input.Runner().Run(ctx, cmd.New().Minikube().Status().SwallowError().Cmd()); err == nil {
		return m.SetContext(ctx, input)
	}
	_ = cmd.New().Minikube().SwallowError().Delete(ctx, input.Runner())
	return cmd.New().Minikube().Cpus(m.Cpus).Memory(m.Memory).VmDriver(m.VmDriver).
		KubeVersion(m.KubeVersion).KubeConfig(input.KubeConfig()).Start(ctx, input.Runner())
}

func (m *Minikube) SetContext(ctx context.Context, input render.InputParams) error {
	return input.Runner().Run(ctx, cmd.New().Kubectl().UseContext(MinikubeContext).Cmd(input.KubeConfig()))
}

func (m *Minikube) Teardown(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(m); err != nil {
		return err
	}
	cmd.Stdout(ctx).Printf("Tearing down minikube cluster")
	return cmd.New().Minikube().KubeConfig(input.KubeConfig()).SwallowError().Delete(ctx, input.Runner())
}
