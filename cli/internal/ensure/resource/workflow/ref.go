package workflow

import (
	"context"
	"gopkg.in/yaml.v2"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

// A Workflow Ref is a path to a file that can be deserialized into a Workflow
type Ref struct {
	RegistryName string        `yaml:"registry" valet:"default=default"`
	Path         string        `yaml:"path"`
	Values       render.Values `yaml:"values"`
	Flags        render.Flags  `yaml:"flags"`
}

func (r *Ref) Load(ctx context.Context, input render.InputParams) (*Workflow, error) {
	input = input.MergeValues(r.Values)
	if err := input.Values.RenderFields(r); err != nil {
		return nil, err
	}
	w, err := r.loadWorkflow(input)
	if err != nil {
		return nil, err
	}
	var filteredSteps []Step
	for _, s := range w.Steps {
		keep := true
		// don't keep resources if a required flag is not set
		for _, requiredFlag := range s.Flags {
			missingRequiredFlag := true
			for _, flag := range input.Flags {
				if flag == requiredFlag {
					missingRequiredFlag = false
					break
				}
			}
			if missingRequiredFlag {
				keep = false
				break
			}
		}
		if keep {
			filteredSteps = append(filteredSteps, s)
		}
	}
	w.Steps = filteredSteps
	return w, nil
}

func (r *Ref) loadWorkflow(input render.InputParams) (*Workflow, error) {
	var w Workflow
	b, err := input.LoadFile(r.RegistryName, r.Path)
	if err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalStrict([]byte(b), &w); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file: %s", err.Error())
		return nil, err
	}
	return &w, nil
}

func (r *Ref) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	workflow, err := r.Load(ctx, input)
	if err != nil {
		return err
	}
	cmd.Stdout().Println("Ensuring workflow %s %s", r.Path, r.Values.ToString())
	if err := workflow.Ensure(ctx, input, command); err != nil {
		return err
	}
	cmd.Stdout().Println("Done ensuring workflow %s", r.Path)
	return nil
}

func (r *Ref) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	workflow, err := r.Load(ctx, input)
	if err != nil {
		return err
	}
	cmd.Stdout().Println("Tearing down workflow %s %s", r.Path, r.Values.ToString())
	if err := workflow.Teardown(ctx, input, command); err != nil {
		return err
	}
	cmd.Stdout().Println("Done tearing down workflow %s", r.Path)
	return nil
}
