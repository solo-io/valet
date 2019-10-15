package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Demos struct {
	Petclinic *Petclinic `yaml:"petclinic"`
}

func (d *Demos) Ensure(ctx context.Context, command cmd.Factory) error {
	if d.Petclinic != nil {
		if err := d.Petclinic.Ensure(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (d *Demos) Teardown(ctx context.Context, command cmd.Factory) error {
	if d.Petclinic != nil {
		if err := d.Petclinic.Teardown(ctx, command); err != nil {
			return err
		}
	}
	return nil
}
