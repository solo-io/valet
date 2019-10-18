package resource

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// A WorkflowRef is a path to a file that can be deserialized into a Workflow
type WorkflowRef struct {
	Path string `yaml:"path"`
}

func (w *WorkflowRef) Ensure(ctx context.Context, command cmd.Factory) error {
	workflow, err := LoadWorkflow(ctx, w.Path)
	if err != nil {
		return err
	}
	return workflow.Ensure(ctx, command)
}

func (w *WorkflowRef) Teardown(ctx context.Context, command cmd.Factory) error {
	workflow, err := LoadWorkflow(ctx, w.Path)
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

func LoadWorkflow(ctx context.Context, path string) (*Workflow, error) {
	var w Workflow

	b, err := loadBytesFromPath(ctx, path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, &w); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to unmarshal file",
			zap.Error(err),
			zap.String("path", path),
			zap.ByteString("bytes", b))
		return nil, err
	}

	return &w, nil
}
