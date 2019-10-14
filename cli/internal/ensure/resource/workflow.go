package resource

import "context"

type Workflow struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`

	URL string `yaml:"url"`
}

func (w *Workflow) Ensure(ctx context.Context) error {
	for _, step := range w.Steps {
		if step.Curl.URL == "" {
			step.Curl.URL = w.URL
		}
		if err := step.Ensure(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Teardown(ctx context.Context) error {
	panic("implement me")
}
