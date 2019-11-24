package application

import (
	"context"
	"fmt"

	"github.com/avast/retry-go"
	"github.com/mitchellh/hashstructure"
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
	Manifests   *Manifests `yaml:"manifests"`
	Application *Ref       `yaml:"application"`

	Values render.Values `yaml:"values"`
	Flags  render.Flags  `yaml:"flags"`
}

func (a *Resource) Ensure(ctx context.Context, input render.InputParams) error {
	if a.Application != nil {
		return a.Application.Ensure(ctx, input.MergeValues(a.Values))
	}
	cmd.Stdout(ctx).Printf("Applying resources")
	applyFunc := func(manifest string) error {
		if manifest == "" {
			return nil
		}
		return cmd.New().Kubectl().ApplyStdIn(ctx, input.Runner(), manifest, input.KubeConfig())
	}
	return a.renderAndRun(ctx, input, applyFunc)
}

func (a *Resource) Teardown(ctx context.Context, input render.InputParams) error {
	if a.Application != nil {
		return a.Application.Teardown(ctx, input.MergeValues(a.Values))
	}
	cmd.Stdout(ctx).Printf("Deleting resources")
	teardownFunc := func(manifest string) error {
		if manifest == "" {
			return nil
		}
		return cmd.New().Kubectl().DeleteStdIn(ctx, input.Runner(), manifest, input.KubeConfig())
	}
	return a.renderAndRun(ctx, input, teardownFunc)
}

func (a *Resource) renderAndRun(ctx context.Context, input render.InputParams, run func(manifest string) error) error {
	manifest, err := a.renderString(ctx, input)
	if err != nil {
		return err
	}
	return retry.Do(func() error {
		return run(manifest)
	}, retry.Attempts(3))
}

func (a *Resource) renderString(ctx context.Context, input render.InputParams) (string, error) {
	rendered, err := a.Render(ctx, input)
	if err != nil {
		return "", err
	}
	if len(rendered) == 0 {
		return "", nil
	}
	hash, err := hashstructure.Hash(rendered, nil)
	if err != nil {
		return "", err
	}
	for _, r := range rendered {
		if r.GetLabels() == nil {
			r.SetLabels(make(map[string]string))
		}
		labels := r.GetLabels()
		labels["valet"] = fmt.Sprintf("%d", hash)
		r.SetLabels(labels)
	}
	manifests, err := helmchart.ManifestsFromResources(rendered)
	if err != nil {
		return "", err
	}
	return manifests.CombinedString(), nil
}

func (a *Resource) Render(ctx context.Context, input render.InputParams) (kuberesource.UnstructuredResources, error) {
	input = input.MergeValues(a.Values)
	return RenderFirst(ctx, input, a.Namespace, a.HelmChart, a.Secret, a.Manifest, a.Manifests, a.Template, a.Application)
}
