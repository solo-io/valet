package application

import (
	"context"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var (
	_ resource.Resource = new(Resource)
	_ Renderable        = new(Resource)
)

type Resource struct {
	Namespace   *Namespace `yaml:"namespace"`
	HelmChart   *HelmChart `yaml:"helmChart"`
	Secret      *Secret    `yaml:"secret"`
	// Deprecated: use Manifest instead
	Path        string     `yaml:"path"`
	Template    *Template  `yaml:"template"`
	Patch       *Patch     `yaml:"patch"`
	Application *Ref       `yaml:"application"`
	Manifest    *Manifest  `yaml:"manifest"`

	Values render.Values `yaml:"values"`
	Flags  render.Flags  `yaml:"flags"`
}

func (a *Resource) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	var manifest *Manifest = nil
	if a.Path != "" {
		manifest = &Manifest{
			Path: a.Path,
		}
	}
	return resource.EnsureFirst(ctx, input, command, a.HelmChart, a.Secret, manifest, a.Template, a.Patch, a.Namespace, a.Application, a.Manifest)
}

func (a *Resource) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	var manifest *Manifest = nil
	if a.Path != "" {
		manifest = &Manifest{
			Path: a.Path,
		}
	}
	return resource.TeardownFirst(ctx, input, command, a.HelmChart, a.Secret, manifest, a.Template, a.Patch, a.Namespace, a.Application, a.Manifest)
}

func (a *Resource) Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	input = input.MergeValues(a.Values)
	var manifest *Manifest = nil
	if a.Path != "" {
		manifest = &Manifest{
			Path: a.Path,
		}
	}
	return RenderFirst(ctx, input, command, a.HelmChart, a.Secret, manifest, a.Template, a.Namespace, a.Application, a.Manifest)
}
