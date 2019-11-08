package workflow

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"gopkg.in/yaml.v2"
)



type Workflow struct {
	Steps          []Step   `yaml:"steps"`
	CleanupSteps   []Step   `yaml:"cleanupSteps"`
	RequiredValues []string `yaml:"requiredValues"`

	Values render.Values `yaml:"values"`
	Flags  render.Flags  `yaml:"flags"`
}

func (w *Workflow) checkRequiredValues(input render.InputParams) error {
	for _, key := range w.RequiredValues {
		if input.Values == nil {
			return render.RequiredValueNotProvidedError(key)
		}
		if _, ok := input.Values[key]; !ok {
			return render.RequiredValueNotProvidedError(key)
		}
	}
	return nil
}

func (w *Workflow) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(w.Values)
	if err := w.checkRequiredValues(input); err != nil {
		return err
	}
	if err := EnsureSteps(ctx, input, command, w.Steps); err != nil {
		return err
	}
	cmd.Stdout().Println("Workflow successful, cleaning up")
	if err := EnsureSteps(ctx, input, command, w.CleanupSteps); err != nil {
		return err
	}
	return nil
}

func EnsureSteps(ctx context.Context, input render.InputParams, command cmd.Factory, steps []Step) error {
	for _, step := range steps {
		if input.Step {
			if err := cmd.PromptPressAnyKeyToContinue(); err != nil {
				return err
			}
		}
		if err := step.Ensure(ctx, input, command); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(w.Values)
	if err := w.checkRequiredValues(input); err != nil {
		return err
	}
	for _, step := range w.Steps {
		if err := step.Teardown(ctx, input, command); err != nil {
			return err
		}
	}
	return nil
}

func LoadWorkflow(path string) (*Workflow, error) {
	var w Workflow

	b, err := render.LoadBytes(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, &w); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file '%s': %s", path, err.Error())
		return nil, err
	}

	return &w, nil
}
