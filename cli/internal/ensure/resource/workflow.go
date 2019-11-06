package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"gopkg.in/yaml.v2"
)

// A WorkflowRef is a path to a file that can be deserialized into a Workflow
type WorkflowRef struct {
	Path string `yaml:"path"`

	Values Values   `yaml:"values"`
	Flags  []string `yaml:"flags"`
}

func (w *WorkflowRef) Load() (*Workflow, error) {
	workflow, err := LoadWorkflow(w.Path)
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

func (w *WorkflowRef) Ensure(ctx context.Context, input InputParams, command cmd.Factory) error {
	input = input.MergeValues(w.Values)
	cmd.Stdout().Println("Ensuring workflow %s %s", w.Path, w.Values.ToString())
	workflow, err := w.Load()
	if err != nil {
		return err
	}
	if err := workflow.Ensure(ctx, input, command); err != nil {
		return err
	}
	cmd.Stdout().Println("Done ensuring workflow %s", w.Path)
	return nil
}

func (w *WorkflowRef) Teardown(ctx context.Context, input InputParams, command cmd.Factory) error {
	input = input.MergeValues(w.Values)
	cmd.Stdout().Println("Tearing down workflow %s %s", w.Path, w.Values.ToString())
	workflow, err := w.Load()
	if err != nil {
		return err
	}
	if err := workflow.Teardown(ctx, input, command); err != nil {
		return err
	}
	cmd.Stdout().Println("Done tearing down workflow %s", w.Path)
	return nil
}

type Workflow struct {
	Steps          []Step   `yaml:"steps"`
	RequiredValues []string `yaml:"requiredValues"`

	Values Values   `yaml:"values"`
	Flags  []string `yaml:"flags"`
}

func (w *Workflow) checkRequiredValues(input InputParams) error {
	for _, key := range w.RequiredValues {
		if input.Values == nil {
			return RequiredValueNotProvidedError(key)
		}
		if _, ok := input.Values[key]; !ok {
			return RequiredValueNotProvidedError(key)
		}
	}
	return nil
}

func (w *Workflow) Ensure(ctx context.Context, input InputParams, command cmd.Factory) error {
	input = input.MergeValues(w.Values)
	if err := w.checkRequiredValues(input); err != nil {
		return err
	}
	for _, step := range w.Steps {
		if err := step.Ensure(ctx, input, command); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Teardown(ctx context.Context, input InputParams, command cmd.Factory) error {
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

	b, err := loadBytesFromPath(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, &w); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file '%s': %s", path, err.Error())
		return nil, err
	}

	return &w, nil
}
