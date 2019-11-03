package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"gopkg.in/yaml.v2"
)

// A WorkflowRef is a path to a file that can be deserialized into a Workflow
type WorkflowRef struct {
	Path string `yaml:"path"`
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
	URL   string `yaml:"url"`
}

func (w *Workflow) Ensure(ctx context.Context, command cmd.Factory) error {
	for _, step := range w.Steps {
		if step.Curl.URL == "" {
			step.Curl.URL = w.URL
		}
		if err := step.Ensure(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Teardown(ctx context.Context, command cmd.Factory) error {
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
