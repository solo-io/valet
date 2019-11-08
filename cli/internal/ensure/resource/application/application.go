package application

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"gopkg.in/yaml.v2"
)

const (
	InstallationStepLabel = "valet.solo.io/installation_step"
)

type Application struct {
	Name           string        `yaml:"name"`
	Resources      []Resource    `yaml:"resources"`
	RequiredValues []string      `yaml:"requiredValues"`
	Values         render.Values `yaml:"values"`
}

func (a *Application) checkRequiredValues(input render.InputParams) error {
	for _, key := range a.RequiredValues {
		if input.Values == nil {
			return render.RequiredValueNotProvidedError(key)
		}
		if _, ok := input.Values[key]; !ok {
			return render.RequiredValueNotProvidedError(key)
		}
	}
	return nil
}

func (a *Application) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	if err := a.checkRequiredValues(input); err != nil {
		return err
	}
	var resources []resource.Resource
	for i := len(a.Resources) - 1; i >= 0; i-- {
		r := &a.Resources[i]
		resources = append(resources, r)
	}
	if err := resource.TeardownAll(ctx, input, command, resources...); err != nil {
		return err
	}
	return nil
}

func (a *Application) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	if err := a.checkRequiredValues(input); err != nil {
		return err
	}
	var resources []resource.Resource
	for _, resource := range a.Resources {
		r := resource
		resources = append(resources, &r)
	}
	if err := resource.EnsureAll(ctx, input, command, resources...); err != nil {
		return err
	}
	return nil
}

func (a *Application) Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	input = input.MergeValues(a.Values)
	if err := a.checkRequiredValues(input); err != nil {
		return nil, err
	}
	var allResources kuberesource.UnstructuredResources
	for i, appResource := range a.Resources {
		if appResource.Patch != nil {
			continue
		}
		renderedResource, err := appResource.Render(ctx, input, command)
		if err != nil {
			return nil, err
		}

		for _, unstructuredResource := range renderedResource {
			labels := unstructuredResource.GetLabels()
			if labels == nil {
				labels = make(map[string]string)
			}
			labels[InstallationStepLabel] = fmt.Sprintf("valet.%s.%d", a.Name, i)
			unstructuredResource.SetLabels(labels)
			allResources = append(allResources, unstructuredResource)
		}
	}
	return allResources, nil
}

func LoadApplication(path string) (*Application, error) {
	var a Application

	b, err := render.LoadBytes(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, &a); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file: %s", err.Error())
		return nil, err
	}

	return &a, nil
}
