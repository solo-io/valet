package workflow

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

type Workflow struct {
	SetupSteps []*Step       `json:"setup,omitempty"`
	Steps      []*Step       `json:"steps,omitempty"`
	Values     render.Values `json:"values,omitempty"`
}

func (w *Workflow) Setup(ctx *api.WorkflowContext) error {
	cmd.Stdout().Println("Setting up workflow")
	for _, step := range w.SetupSteps {
		knownStep := step.Get()
		values := w.Values
		if values == nil && step.Values != nil {
			values = make(map[string]string)
		}
		values = values.MergeValues(step.Values)
		description, err := knownStep.GetDescription(ctx, values)
		if err != nil {
			return err
		}
		cmd.Stdout().Println(description)
		if err := knownStep.Run(ctx, values); err != nil {
			return err
		}
	}
	cmd.Stdout().Println("Workflow setup successfully")
	return nil
}

func (w *Workflow) Run(ctx *api.WorkflowContext) error {
	cmd.Stdout().Println("Running workflow")
	for _, step := range w.Steps {
		knownStep := step.Get()
		values := w.Values
		if values == nil && step.Values != nil {
			values = make(map[string]string)
		}
		values = values.MergeValues(step.Values)
		description, err := knownStep.GetDescription(ctx, values)
		if err != nil {
			return err
		}
		cmd.Stdout().Println(description)
		if err := knownStep.Run(ctx, values); err != nil {
			return err
		}
	}
	cmd.Stdout().Println("Workflow finished successfully")
	return nil
}
