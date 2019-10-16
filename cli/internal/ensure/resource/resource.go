package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Resource interface {
	Ensure(ctx context.Context, command cmd.Factory) error
	Teardown(ctx context.Context, command cmd.Factory) error
}

type Resources struct {
	Paths []string
}

func (r *Resources) Ensure(ctx context.Context, command cmd.Factory) error {
	for _, path := range r.Paths {
		if err := command.Kubectl().ApplyFile(path).Cmd().Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resources) Teardown(ctx context.Context, command cmd.Factory) error {
	for _, path := range r.Paths {
		if err := command.Kubectl().DeleteFile(path).IgnoreNotFound().Cmd().Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
