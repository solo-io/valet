package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"gopkg.in/yaml.v2"
)

type Application struct {
	Name           string                `yaml:"name"`
	Resources      []ApplicationResource `yaml:"resources"`
	RequiredValues []string              `yaml:"requiredValues"`
	Values         Values                `yaml:"values"`
}

type ApplicationRef struct {
	Path   string `yaml:"path" valet:"template"`
	Values Values `yaml:"values"`
	Flags  Flags  `yaml:"flags"`
}

func (a *ApplicationRef) Load(ctx context.Context, input InputParams) (*Application, error) {
	app, err := LoadApplication(a.Path)
	if err != nil {
		return nil, err
	}
	var filteredResources []ApplicationResource
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

func (a *ApplicationRef) Ensure(ctx context.Context, input InputParams, command cmd.Factory) error {
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

func (a *ApplicationRef) Teardown(ctx context.Context, input InputParams, command cmd.Factory) error {
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

func (a *Application) checkRequiredValues(input InputParams) error {
	for _, key := range a.RequiredValues {
		if input.Values == nil {
			return RequiredValueNotProvidedError(key)
		}
		if _, ok := input.Values[key]; !ok {
			return RequiredValueNotProvidedError(key)
		}
	}
	return nil
}

func (a *Application) Teardown(ctx context.Context, input InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	if err := a.checkRequiredValues(input); err != nil {
		return err
	}
	var resources []Resource
	for i := len(a.Resources) - 1; i >= 0; i-- {
		r := &a.Resources[i]
		resources = append(resources, r)
	}
	if err := TeardownAll(ctx, input, command, resources...); err != nil {
		return err
	}
	return nil
}

func (a *Application) Ensure(ctx context.Context, input InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	if err := a.checkRequiredValues(input); err != nil {
		return err
	}
	var resources []Resource
	for _, resource := range a.Resources {
		r := resource
		resources = append(resources, &r)
	}
	if err := EnsureAll(ctx, input, command, resources...); err != nil {
		return err
	}
	return nil
}

func LoadApplication(path string) (*Application, error) {
	var a Application

	b, err := loadBytesFromPath(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, &a); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file: %s", err.Error())
		return nil, err
	}

	return &a, nil
}
