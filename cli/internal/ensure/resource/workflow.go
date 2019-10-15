package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Workflow struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`

	URL string `yaml:"url"`
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
	panic("implement me")
}
