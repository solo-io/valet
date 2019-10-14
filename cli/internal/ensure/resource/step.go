package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Step struct {
	Apply  string `yaml:"apply"`
	Delete string `yaml:"delete"`
	Curl   *Curl  `yaml:"curl"`
}

func (s *Step) Ensure(ctx context.Context) error {
	if s.Apply != "" {
		if err := cmd.Kubectl().ApplyFile(s.Apply).Run(ctx); err != nil {
			return err
		}
	}
	if s.Delete != "" {
		if err := cmd.Kubectl().DeleteFile(s.Delete).Run(ctx); err != nil {
			return err
		}
	}
	if s.Curl != nil {
		return s.Curl.Ensure(ctx)
	}
	return nil
}

func (s *Step) Teardown(ctx context.Context) error {
	return nil
}


