package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Resource interface {
	Ensure(ctx context.Context, command cmd.Factory) error
	Teardown(ctx context.Context, command cmd.Factory) error
}

var (
	_ Resource = new(Config)

	_ Resource = new(Cluster)
	_ Resource = new(GKE)
	_ Resource = new(Minikube)

	_ Resource = new(Application)
	_ Resource = new(ApplicationResource)
	_ Resource = new(HelmChart)
	_ Resource = new(Manifest)
	_ Resource = new(Secret)
	_ Resource = new(Namespace)
	_ Resource = new(Template)
	_ Resource = new(Patch)
	_ Resource = new(DnsEntry)
	_ Resource = new(Condition)

	_ Resource = new(Workflow)
	_ Resource = new(Curl)
	_ Resource = new(Step)
)

type ApplicationResource struct {
	Namespace   *Namespace      `yaml:"namespace"`
	HelmChart   *HelmChart      `yaml:"helmChart"`
	Secret      *Secret         `yaml:"secret"`
	Path        string          `yaml:"path"`
	Template    *Template       `yaml:"template"`
	Patch       *Patch          `yaml:"patch"`
	Application *ApplicationRef `yaml:"application"`

	Values Values   `yaml:"values"`
	Flags  []string `yaml:"flags"`
}

func (a *ApplicationResource) updateWithValues(values Values) {
	a.Values = MergeValues(a.Values, values)
}

func (a *ApplicationResource) Ensure(ctx context.Context, command cmd.Factory) error {
	if a.HelmChart != nil {
		if err := a.HelmChart.updateWithValues(a.Values); err != nil {
			return err
		}
		return a.HelmChart.Ensure(ctx, command)
	}
	if a.Secret != nil {
		if err := a.Secret.updateWithValues(a.Values); err != nil {
			return err
		}
		return a.Secret.Ensure(ctx, command)
	}
	if a.Path != "" {
		manifest := Manifest{
			Path: a.Path,
		}
		return manifest.Ensure(ctx, command)
	}
	if a.Template != nil {
		a.Template.updateWithValues(a.Values)
		return a.Template.Ensure(ctx, command)
	}
	if a.Patch != nil {
		a.Patch.updateWithValues(a.Values)
		return a.Patch.Ensure(ctx, command)
	}
	if a.Namespace != nil {
		return a.Namespace.Ensure(ctx, command)
	}
	if a.Application != nil {
		a.Application.updateWithValues(a.Values)
		return a.Application.Ensure(ctx, command)
	}
	return nil
}

func (a *ApplicationResource) Teardown(ctx context.Context, command cmd.Factory) error {
	if a.HelmChart != nil {
		if err := a.HelmChart.updateWithValues(a.Values); err != nil {
			return err
		}
		return a.HelmChart.Teardown(ctx, command)
	}
	if a.Secret != nil {
		if err := a.Secret.updateWithValues(a.Values); err != nil {
			return err
		}
		return a.Secret.Teardown(ctx, command)
	}
	if a.Path != "" {
		manifest := Manifest{
			Path: a.Path,
		}
		return manifest.Teardown(ctx, command)
	}
	if a.Template != nil {
		a.Template.updateWithValues(a.Values)
		return a.Template.Teardown(ctx, command)
	}
	if a.Patch != nil {
		a.Patch.updateWithValues(a.Values)
		return a.Patch.Teardown(ctx, command)
	}
	if a.Namespace != nil {
		return a.Namespace.Teardown(ctx, command)
	}
	if a.Application != nil {
		a.Application.updateWithValues(a.Values)
		return a.Application.Teardown(ctx, command)
	}
	return nil
}
