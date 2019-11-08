package application

import (
	"context"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var (
	_ resource.Resource = new(Ref)
	_ Renderable        = new(Ref)
)

type Ref struct {
	Path   string        `yaml:"path" valet:"template"`
	Values render.Values `yaml:"values"`
	Flags  render.Flags  `yaml:"flags"`
}

func (a *Ref) Load(ctx context.Context, input render.InputParams) (*Application, error) {
	app, err := LoadApplication(a.Path)
	if err != nil {
		return nil, err
	}
	var filteredResources []Resource
	for _, resource := range app.Resources {
		keep := true
		// don't keep resources if a required flag is not set
		for _, requiredFlag := range resource.Flags {
			missingRequiredFlag := true
			for _, flag := range input.Flags {
				if flag == requiredFlag {
					missingRequiredFlag = false
					break
				}
			}
			if missingRequiredFlag {
				keep = false
				break
			}
		}
		if keep {
			filteredResources = append(filteredResources, resource)
		}
	}
	app.Resources = filteredResources
	return app, nil
}

func (a *Ref) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	if err := input.Values.RenderFields(a); err != nil {
		return err
	}
	cmd.Stdout().Println("Ensuring application %s values=%s flags=%s", a.Path, input.Values.ToString(), input.Flags.ToString())
	app, err := a.Load(ctx, input)
	if err != nil {
		return err
	}
	err = app.Ensure(ctx, input, command)
	if err == nil {
		cmd.Stdout().Println("Done ensuring application %s", a.Path)
	}
	return err
}

func (a *Ref) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	if err := input.Values.RenderFields(a); err != nil {
		return err
	}
	app, err := a.Load(ctx, input)
	if err != nil {
		return err
	}
	return app.Teardown(ctx, input, command)
}

func (a *Ref) Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	input = input.MergeValues(a.Values)
	if err := input.Values.RenderFields(a); err != nil {
		return nil, err
	}
	app, err := a.Load(ctx, input)
	if err != nil {
		return nil, err
	}
	return app.Render(ctx, input, command)
}
