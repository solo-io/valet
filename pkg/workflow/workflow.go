package workflow

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

type Workflow struct {
	Steps  []*Step       `json:"steps,omitempty"`
	Values render.Values `json:"values,omitempty"`
}

func (w *Workflow) Run(ctx *api.WorkflowContext) error {
	for _, step := range w.Steps {
		knownStep := step.Get()
		description, err := knownStep.GetDescription(ctx, w.Values)
		if err != nil {
			return err
		}
		cmd.Stdout().Println(description)
		if err := knownStep.Run(ctx, step.Values); err != nil {
			return err
		}
	}
	cmd.Stdout().Println("Workflow finished successfully")
	return nil
}
