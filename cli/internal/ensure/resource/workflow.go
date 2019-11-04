package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"gopkg.in/yaml.v2"
)

// A WorkflowRef is a path to a file that can be deserialized into a Workflow
type WorkflowRef struct {
	Path string `yaml:"path"`

	Values map[string]string `yaml:"values"`
	Flags  []string          `yaml:"flags"`
}

func (w *WorkflowRef) updateWithValues(values map[string]string) {
	for k, v := range values {
		if w.Values == nil {
			w.Values = make(map[string]string)
		}
		w.Values[k] = v
	}
}

func (w *WorkflowRef) Load() (*Workflow, error) {
	workflow, err := LoadWorkflow(w.Path)
	if err != nil {
		return nil, err
	}
	for k, v := range w.Values {
		if workflow.Values == nil {
			workflow.Values = make(map[string]string)
		}
		workflow.Values[k] = v
		for _, step := range workflow.Steps {
			if step.Values == nil {
				step.Values = make(map[string]string)
			}
			step.Values[k] = v
			if step.Install != nil {
				step.Install.Flags = w.Flags
				step.Install.updateWithValues(w.Values)
			}
			if step.Uninstall != nil {
				step.Uninstall.Flags = w.Flags
				step.Uninstall.updateWithValues(w.Values)
			}
		}
	}
	return workflow, nil
}

func (w *WorkflowRef) Ensure(ctx context.Context, command cmd.Factory) error {
	workflow, err := LoadWorkflow(w.Path)
	if err != nil {
		return err
	}
	return workflow.Ensure(ctx, command)
}

func (w *WorkflowRef) Teardown(ctx context.Context, command cmd.Factory) error {
	workflow, err := LoadWorkflow(w.Path)
	if err != nil {
		return err
	}
	return workflow.Teardown(ctx, command)
}

type Workflow struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`

	Values map[string]string `yaml:"values"`
	Flags  []string          `yaml:"flags"`
}

func (w *Workflow) Ensure(ctx context.Context, command cmd.Factory) error {
	for _, step := range w.Steps {
		if err := step.Ensure(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Teardown(ctx context.Context, command cmd.Factory) error {
	for _, step := range w.Steps {
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
