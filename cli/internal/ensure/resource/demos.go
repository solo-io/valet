package resource

import "context"

type Demos struct {
	Petclinic *Petclinic `yaml:"petclinic"`

	Glooctl *Glooctl
}

func (d *Demos) Ensure(ctx context.Context) error {
	if d.Petclinic != nil {
		if d.Glooctl != nil {
			d.Petclinic.Glooctl = d.Glooctl
		}
		if err := d.Petclinic.Ensure(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (d *Demos) Teardown(ctx context.Context) error {
	if d.Petclinic != nil {
		if err := d.Petclinic.Teardown(ctx); err != nil {
			return err
		}
	}
	return nil
}
