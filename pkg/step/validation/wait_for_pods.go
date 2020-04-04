package validation

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"
)

type WaitForPods struct {
	Namespace string `json:"namespace"`
}

func (w *WaitForPods) GetDescription(_ *api.WorkflowContext, _ render.Values) (string, error) {
	return fmt.Sprintf("Waiting for pods in namespace %s", w.Namespace), nil
}

func (w *WaitForPods) Run(ctx *api.WorkflowContext, values render.Values) error {
	return ctx.KubeClient.WaitUntilPodsRunning(w.Namespace)
}

func (w *WaitForPods) GetDocs(ctx *api.WorkflowContext, options api.DocsOptions) (string, error) {
	panic("implement me")
}
