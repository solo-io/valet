package check

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"
)

// check.WaitForPods is a workflow step that is used to pause a workflow until
// the pods in a namespace are ready or completed successfully.
type WaitForPods struct {
	// Namespace to check for pods
	Namespace string `json:"namespace,omitempty"`
}

func (w *WaitForPods) GetDescription(_ *api.WorkflowContext, _ render.Values) (string, error) {
	return fmt.Sprintf("Waiting for pods in namespace %s", w.Namespace), nil
}

func (w *WaitForPods) Run(ctx *api.WorkflowContext, _ render.Values) error {
	return ctx.KubeClient.WaitUntilPodsRunning(w.Namespace)
}

func (w *WaitForPods) GetDocs(_ *api.WorkflowContext, _ render.Values, _ render.Flags) (string, error) {
	return fmt.Sprintf("Wait until the pods in namespace '%s' are ready. Use `kubectl get pods -n %s` to check the status.", w.Namespace, w.Namespace), nil
}
