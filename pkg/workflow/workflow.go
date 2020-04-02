package workflow

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
)

type Workflow struct {
	Steps []Step `json:"steps"`
}

func (w *Workflow) Run(ctx *api.WorkflowContext) error {
	for _, step := range w.Steps {
		knownStep := step.Get()
		cmd.Stdout().Println(knownStep.GetDescription())
		if err := knownStep.Run(ctx, step.Values); err != nil {
			return err
		}
	}
	cmd.Stdout().Println("Workflow finished successfully")
	return nil
}
