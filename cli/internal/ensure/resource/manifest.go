package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ Resource = new(Manifest)

type Manifest struct {
	Path string `yaml:"path"`
}

func (m *Manifest) Ensure(ctx context.Context) error {
	return cmd.Kubectl().Apply().File(m.Path).Run()
}

func (m *Manifest) Teardown(ctx context.Context) error {
	return cmd.Kubectl().DeleteFile(m.Path).Run()
}
