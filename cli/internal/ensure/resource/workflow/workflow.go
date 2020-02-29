package workflow

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Workflow struct {
	Steps          []Step   `yaml:"steps"`
	CleanupSteps   []Step   `yaml:"cleanupSteps"`
	RequiredValues []string `yaml:"requiredValues"`

	Values render.Values `yaml:"values"`
	Flags  render.Flags  `yaml:"flags"`

	Docs *Docs `yaml:"docs"`
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

func (w *Workflow) Ensure(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(w.Values)
	if err := w.checkRequiredValues(input); err != nil {
		return err
	}
	if err := EnsureSteps(ctx, input, w.Steps); err != nil {
		return err
	}
	cmd.Stdout().Println("Workflow successful, cleaning up")
	if err := EnsureSteps(ctx, input, w.CleanupSteps); err != nil {
		return err
	}
	return nil
}

func EnsureSteps(ctx context.Context, input render.InputParams, steps []Step) error {
	for _, step := range steps {
		if input.Step {
			if err := cmd.PromptPressAnyKeyToContinue(); err != nil {
				return err
			}
		}
		if err := step.Ensure(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Teardown(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(w.Values)
	if err := w.checkRequiredValues(input); err != nil {
		return err
	}
	for _, step := range w.Steps {
		if err := step.Teardown(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Document(ctx context.Context, input render.InputParams, section *Section) {
	if w.Docs != nil {
		section.Title = w.Docs.Title
		section.Description = w.Docs.Description
		section.Notes = w.Docs.Notes
	}

	for _, step := range w.Steps {
		stepSection := Section{}
		step.Document(ctx, input, &stepSection)
		section.Sections = append(section.Sections, stepSection)
	}
}
