package cluster

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

const (
	MinikubeContext = "minikube"
)

var _ ClusterResource = new(Minikube)

type Minikube struct {
	Cpus         int      `json:"cpus" valet:"template,key=MinikubeCpus,default=4"`
	Memory       int      `json:"memory" valet:"template,key=MinikubeMemory,default=8192"`
	KubeVersion  string   `json:"version" valet:"template,key=KubeVersion,default=v1.13.0"`
	VmDriver     string   `json:"vmDriver" valet:"template,key=MinikubeVmDriver,default=virtualbox"`
	FeatureGates []string `json:"featureGates"`
}

func (m *Minikube) Ensure(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(m, ctx.Runner); err != nil {
		return err
	}
	cmd.Stdout().Println("Ensuring minikube cluster (cpus: %d, memory: %d, version: %s, driver: %s, featureGates: %v)",
		m.Cpus, m.Memory, m.KubeVersion, m.VmDriver, m.FeatureGates)
	// If minikube status seems healthy, just set context and return
	if err := ctx.Runner.Run(cmd.New().Minikube().Status().SwallowError().Cmd()); err == nil {
		return m.SetContext(ctx, values)
	}
	_ = cmd.New().Minikube().SwallowError().Delete(ctx.Runner)

	return cmd.New().Minikube().Cpus(m.Cpus).Memory(m.Memory).VmDriver(m.VmDriver).
		KubeVersion(m.KubeVersion).FeatureGates(m.FeatureGates).Start(ctx.Runner)
}

func (m *Minikube) SetContext(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(m, ctx.Runner); err != nil {
		return err
	}
	return ctx.Runner.Run(cmd.New().Kubectl().UseContext(MinikubeContext).Cmd())
}

func (m *Minikube) Teardown(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(m, ctx.Runner); err != nil {
		return err
	}
	cmd.Stdout().Println("Tearing down minikube cluster")
	return cmd.New().Minikube().SwallowError().Delete(ctx.Runner)
}
