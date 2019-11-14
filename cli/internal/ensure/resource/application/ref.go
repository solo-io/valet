package application

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"

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
	RegistryName string        `yaml:"registry" valet:"default=default"`
	Path         string        `yaml:"path" valet:"template"`
	Values       render.Values `yaml:"values"`
	Flags        render.Flags  `yaml:"flags"`
}

func (a *Ref) Load(ctx context.Context, input render.InputParams) (*Application, error) {
	input = input.MergeValues(a.Values)
	if err := input.RenderFields(a); err != nil {
		return nil, err
	}
	app, err := a.loadApplication(input)
	if err != nil {
		return nil, err
	}
	var filteredResources []Resource
	for _, r := range app.Resources {
		keep := true
		// don't keep resources if a required flag is not set
		for _, requiredFlag := range r.Flags {
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
			filteredResources = append(filteredResources, r)
		}
	}
	app.Resources = filteredResources
	return app, nil
}

func (a *Ref) loadApplication(input render.InputParams) (*Application, error) {
	var app Application
	b, err := input.LoadFile(a.RegistryName, a.Path)
	if err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalStrict([]byte(b), &app); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file: %s", err.Error())
		return nil, err
	}
	return &app, nil
}

func (a *Ref) Ensure(ctx context.Context, input render.InputParams) error {
	app, err := a.Load(ctx, input)
	if err != nil {
		return err
	}
	input = input.MergeValues(a.Values)
	appRegistry, err := input.GetRegistry(a.RegistryName)
	if err != nil {
		return err
	}
	input.SetRegistry(render.DefaultRegistry, appRegistry)
	appString := a.Path
	if a.RegistryName != render.DefaultRegistry {
		appString = fmt.Sprintf("%s:%s", a.RegistryName, appString)
	}
	cmd.Stdout().Println("Ensuring application %s values=%s flags=%s", appString, input.Values.ToString(), input.Flags.ToString())
	err = app.Ensure(ctx, input)
	if err == nil {
		cmd.Stdout().Println("Done ensuring application %s", a.Path)
	}
	return err
}

func (a *Ref) Teardown(ctx context.Context, input render.InputParams) error {
	app, err := a.Load(ctx, input)
	if err != nil {
		return err
	}
	input = input.MergeValues(a.Values)
	appRegistry, err := input.GetRegistry(a.RegistryName)
	if err != nil {
		return err
	}
	input.SetRegistry(render.DefaultRegistry, appRegistry)
	cmd.Stdout().Println("Tearing down application %s values=%s flags=%s", a.Path, input.Values.ToString(), input.Flags.ToString())
	err = app.Teardown(ctx, input)
	if err == nil {
		cmd.Stdout().Println("Done tearing down application %s", a.Path)
	}
	return err
}

func (a *Ref) Render(ctx context.Context, input render.InputParams) (kuberesource.UnstructuredResources, error) {
	app, err := a.Load(ctx, input)
	if err != nil {
		return nil, err
	}
	input = input.MergeValues(a.Values)
	appRegistry, err := input.GetRegistry(a.RegistryName)
	if err != nil {
		return nil, err
	}
	input.SetRegistry(render.DefaultRegistry, appRegistry)
	return app.Render(ctx, input)
}
