package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ Resource = new(Manifest)

type Manifest struct {
	Path string `yaml:"path"`
}

func (m *Manifest) Ensure(ctx context.Context, command cmd.Factory) error {
	return command.Kubectl().Apply().File(m.Path).Run(ctx)
}

func (m *Manifest) Teardown(ctx context.Context, command cmd.Factory) error {
	return command.Kubectl().DeleteFile(m.Path).Run(ctx)
}
