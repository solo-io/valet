package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Resource interface {
	Ensure(ctx context.Context) error
	Teardown(ctx context.Context) error
}

type Resources struct {
	Paths []string
}

func (r *Resources) Ensure(ctx context.Context) error {
	return cmd.KubectlApplyAllFiles(ctx, r.Paths)
}

func (r *Resources) Teardown(ctx context.Context) error {
	return cmd.KubectlDeleteAllFiles(ctx, r.Paths)
}


