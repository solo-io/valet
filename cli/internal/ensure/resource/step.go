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

func (s *Step) Ensure(ctx context.Context, command cmd.Factory) error {
	if s.Apply != "" {
		if err := command.Kubectl().ApplyFile(s.Apply).Cmd().Run(ctx); err != nil {
			return err
		}
	}
	if s.Delete != "" {
		if err := command.Kubectl().DeleteFile(s.Delete).Cmd().Run(ctx); err != nil {
			return err
		}
	}
	if s.Curl != nil {
		return s.Curl.Ensure(ctx, command)
	}
	return nil
}

func (s *Step) Teardown(ctx context.Context, command cmd.Factory) error {
	return nil
}


