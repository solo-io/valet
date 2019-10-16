package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Demos struct {
	Petclinic *Petclinic `yaml:"petclinic"`
	Petstore  *Petstore  `yaml:"petstore"`
}

func (d *Demos) Ensure(ctx context.Context, command cmd.Factory) error {
	return EnsureAll(ctx, command, d.Petclinic, d.Petstore)
}

func (d *Demos) Teardown(ctx context.Context, command cmd.Factory) error {
	return TeardownAll(ctx, command, d.Petclinic, d.Petstore)
}
