package workflow

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Workflow struct {
	Docs

	Steps          []Step   `json:"steps"`
	CleanupSteps   []Step   `json:"cleanupSteps"`
	RequiredValues []string `json:"requiredValues"`

	Values render.Values `json:"values"`
	Flags  render.Flags  `json:"flags"`
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

func (w *Workflow) Document(ctx context.Context, input render.InputParams, section *Section) error {
	section.Title = w.Title
	section.Description = w.Description
	section.Notes = w.Notes

	for _, step := range w.Steps {
		stepSection := Section{}
		err := step.Document(ctx, input, &stepSection)
		if err != nil {
			return err
		}
		section.Sections = append(section.Sections, stepSection)
	}
	return nil
}
