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
	Path   string `yaml:"path"`
	Values Values `yaml:"values"`
	Flags  Flags  `yaml:"flags"`
}

func (a *ApplicationRef) updateWithValues(values Values) error {
	a.Values = MergeValues(a.Values, values)
	path, err := LoadTemplate(a.Path, values)
	if err != nil {
		return err
	}
	a.Path = path
	return nil
}

func (a *ApplicationRef) updateWithFlags(flags Flags) {
	a.Flags = append(a.Flags, flags...)
}

func (a *ApplicationRef) Load(ctx context.Context) (*Application, error) {
	app, err := LoadApplication(a.Path)
	if err != nil {
		return nil, err
	}
	app.updateWithValues(a.Values)
	var filteredResources []ApplicationResource
	for _, resource := range app.Resources {
		keep := true
		// don't keep resources if a required flag is not set
		for _, requiredFlag := range resource.Flags {
			missingRequiredFlag := true
			for _, flag := range a.Flags {
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
			if resource.Application != nil {
				// pass on the runtime flags, which at this point we know to be a superset of the required flags
				resource.Application.Flags = a.Flags
			}
		}
	}
	app.Resources = filteredResources
	return app, nil
}

func (a *ApplicationRef) Ensure(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Ensuring application %s values=%s flags=%s", a.Path, a.Values.ToString(), a.Flags.ToString())
	app, err := a.Load(ctx)
	if err != nil {
		return err
	}
	err = app.Ensure(ctx, command)
	if err == nil {
		cmd.Stdout().Println("Done ensuring application %s", a.Path)
	}
	return err
}

func (a *ApplicationRef) Teardown(ctx context.Context, command cmd.Factory) error {
	app, err := a.Load(ctx)
	if err != nil {
		return err
	}
	return app.Teardown(ctx, command)
}

func (a *Application) updateWithValues(values Values) {
	a.Values = MergeValues(a.Values, values)
}

func (a *Application) checkRequiredValues() error {
	for _, key := range a.RequiredValues {
		if a.Values == nil {
			return RequiredValueNotProvidedError(key)
		}
		if _, ok := a.Values[key]; !ok {
			return RequiredValueNotProvidedError(key)
		}
	}
	return nil
}

func (a *Application) Teardown(ctx context.Context, command cmd.Factory) error {
	if err := a.checkRequiredValues(); err != nil {
		return err
	}
	var resources []Resource
	for i := len(a.Resources) - 1; i >= 0; i-- {
		r := &a.Resources[i]
		r.updateWithValues(a.Values)
		resources = append(resources, r)
	}
	if err := TeardownAll(ctx, command, resources...); err != nil {
		return err
	}
	return nil
}

func (a *Application) Ensure(ctx context.Context, command cmd.Factory) error {
	if err := a.checkRequiredValues(); err != nil {
		return err
	}
	var resources []Resource
	for _, resource := range a.Resources {
		r := resource
		r.updateWithValues(a.Values)
		resources = append(resources, &r)
	}
	if err := EnsureAll(ctx, command, resources...); err != nil {
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
