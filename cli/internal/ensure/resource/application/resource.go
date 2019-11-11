package application

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/solo-io/go-utils/installutils/helmchart"

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
	Template    *Template  `yaml:"template"`
	Manifest    *Manifest  `yaml:"manifest"`
	Application *Ref       `yaml:"application"`

	Values render.Values `yaml:"values"`
	Flags  render.Flags  `yaml:"flags"`
}

func (a *Resource) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	manifest, err := a.renderString(ctx, input, command)
	if err != nil {
		return err
	}
	return retry.Do(func() error {
		return command.Kubectl().ApplyStdIn(manifest).Cmd().Run(ctx)
	}, retry.Attempts(3))
}

func (a *Resource) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	manifest, err := a.renderString(ctx, input, command)
	if err != nil {
		return err
	}
	return retry.Do(func() error {
		return command.Kubectl().DeleteStdIn(manifest).IgnoreNotFound().Cmd().Run(ctx)
	}, retry.Attempts(3))
}

func (a *Resource) renderString(ctx context.Context, input render.InputParams, command cmd.Factory) (string, error) {
	rendered, err := a.Render(ctx, input, command)
	if err != nil {
		return "", err
	}
	manifests, err := helmchart.ManifestsFromResources(rendered)
	if err != nil {
		return "", err
	}
	return manifests.CombinedString(), nil
}

func (a *Resource) Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	input = input.MergeValues(a.Values)
	return RenderFirst(ctx, input, command, a.Namespace, a.HelmChart, a.Secret, a.Manifest, a.Application)
}
