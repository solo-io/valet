package cluster

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	MinikubeContext = "minikube"
)

var _ ClusterResource = new(Minikube)

type Minikube struct {
	Cpus        int    `json:"cpus" valet:"template,key=MinikubeCpus,default=4"`
	Memory      int    `json:"memory" valet:"template,key=MinikubeMemory,default=8192"`
	KubeVersion string `json:"name" valet:"template,key=KubeVersion,default=v1.13.0"`
	VmDriver    string `json:"vmDriver" valet:"template,key=MinikubeVmDriver,default=virtualbox"`
}

func (m *Minikube) Ensure(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(m); err != nil {
		return err
	}
	cmd.Stdout().Println("Ensuring minikube cluster (cpus: %d, memory: %d, version: %s, driver: %s)",
		m.Cpus, m.Memory, m.KubeVersion, m.VmDriver)
	// If minikube status seems healthy, just set context and return
	if err := input.Runner().Run(ctx, cmd.New().Minikube().Status().SwallowError().Cmd()); err == nil {
		return m.SetContext(ctx, input.Runner())
	}
	_ = cmd.New().Minikube().SwallowError().Delete(ctx, input.Runner())
	return cmd.New().Minikube().Cpus(m.Cpus).Memory(m.Memory).VmDriver(m.VmDriver).
		KubeVersion(m.KubeVersion).Start(ctx, input.Runner())
}

func (m *Minikube) SetContext(ctx context.Context, runner cmd.Runner) error {
	return runner.Run(ctx, cmd.New().Kubectl().UseContext(MinikubeContext).Cmd())
}

func (m *Minikube) Teardown(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(m); err != nil {
		return err
	}
	cmd.Stdout().Println("Tearing down minikube cluster")
	return cmd.New().Minikube().SwallowError().Delete(ctx, input.Runner())
}
