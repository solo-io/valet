package workflow

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

// A Workflow Ref is a path to a file that can be deserialized into a Workflow
type Ref struct {
	Path string `yaml:"path"`

	Values render.Values `yaml:"values"`
	Flags  render.Flags  `yaml:"flags"`
}

func (r *Ref) Load() (*Workflow, error) {
	workflow, err := LoadWorkflow(r.Path)
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

func (r *Ref) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(r.Values)
	cmd.Stdout().Println("Ensuring workflow %s %s", r.Path, r.Values.ToString())
	workflow, err := r.Load()
	if err != nil {
		return err
	}
	if err := workflow.Ensure(ctx, input, command); err != nil {
		return err
	}
	cmd.Stdout().Println("Done ensuring workflow %s", r.Path)
	return nil
}

func (r *Ref) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(r.Values)
	cmd.Stdout().Println("Tearing down workflow %s %s", r.Path, r.Values.ToString())
	workflow, err := r.Load()
	if err != nil {
		return err
	}
	if err := workflow.Teardown(ctx, input, command); err != nil {
		return err
	}
	cmd.Stdout().Println("Done tearing down workflow %s", r.Path)
	return nil
}
