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

func (w *WorkflowRef) updateWithValues(values Values) {
	w.Values = MergeValues(w.Values, values)
}

func (w *WorkflowRef) updateWithFlags(flags []string) {
	w.Flags = append(w.Flags, flags...)
}

func (w *WorkflowRef) Load() (*Workflow, error) {
	workflow, err := LoadWorkflow(w.Path)
	if err != nil {
		return nil, err
	}
	workflow.updateWithValues(w.Values)
	workflow.updateWithFlags(w.Flags)
	return workflow, nil
}

func (w *WorkflowRef) Ensure(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Ensuring workflow %s %s", w.Path, w.Values.ToString())
	workflow, err := w.Load()
	if err != nil {
		return err
	}
	if err := workflow.Ensure(ctx, command); err != nil {
		return err
	}
	cmd.Stdout().Println("Done ensuring workflow %s", w.Path)
	return nil
}

func (w *WorkflowRef) Teardown(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Tearing down workflow %s %s", w.Path, w.Values.ToString())
	workflow, err := w.Load()
	if err != nil {
		return err
	}
	if err := workflow.Teardown(ctx, command); err != nil {
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

func (w *Workflow) checkRequiredValues() error {
	for _, key := range w.RequiredValues {
		if w.Values == nil {
			return RequiredValueNotProvidedError(key)
		}
		if _, ok := w.Values[key]; !ok {
			return RequiredValueNotProvidedError(key)
		}
	}
	return nil
}

func (w *Workflow) updateWithValues(values Values) {
	w.Values = MergeValues(w.Values, values)
}

func (w *Workflow) updateWithFlags(flags []string) {
	w.Flags = flags
}

func (w *Workflow) Ensure(ctx context.Context, command cmd.Factory) error {
	if err := w.checkRequiredValues(); err != nil {
		return err
	}
	for _, step := range w.Steps {
		step.updateWithFlags(w.Flags)
		step.updateWithValues(w.Values)
		if err := step.Ensure(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Teardown(ctx context.Context, command cmd.Factory) error {
	if err := w.checkRequiredValues(); err != nil {
		return err
	}
	for _, step := range w.Steps {
		step.updateWithValues(w.Values)
		step.updateWithFlags(w.Flags)
		if err := step.Teardown(ctx, command); err != nil {
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
