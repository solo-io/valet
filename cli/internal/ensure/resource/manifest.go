package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ Resource = new(Manifest)

type Manifest struct {
	Path string `yaml:"path"`
}

func (m *Manifest) Ensure(ctx context.Context, _ InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Ensuring manifest %s", m.Path)
	return command.Kubectl().Apply().File(m.Path).Cmd().Run(ctx)
}

func (m *Manifest) Teardown(ctx context.Context, _ InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Tearing down manifest %s", m.Path)
	return command.Kubectl().DeleteFile(m.Path).IgnoreNotFound().Cmd().Run(ctx)
}
