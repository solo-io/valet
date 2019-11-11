package application

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

const (
	InstallationStepLabel = "valet.solo.io/installation_step"
)

type Application struct {
	Name           string        `yaml:"name"`
	Namespace      string        `yaml:"namespace" valet:"key=Namespace,default=default"`
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
	if err := input.Values.RenderFields(a); err != nil {
		return err
	}
	for i := len(a.Resources) - 1; i >= 0; i-- {
		if err := a.Resources[i].Teardown(ctx, input, command); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	if err := a.checkRequiredValues(input); err != nil {
		return err
	}
	if err := input.Values.RenderFields(a); err != nil {
		return err
	}
	for _, r := range a.Resources {
		if err := r.Ensure(ctx, input, command); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) getLabel(step int) string {
	return fmt.Sprintf("valet.%s.%d", a.Name, step)
}

func (a *Application) Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	input = input.MergeValues(a.Values)
	if err := a.checkRequiredValues(input); err != nil {
		return nil, err
	}
	var allResources kuberesource.UnstructuredResources
	for i, appResource := range a.Resources {
		renderedResource, err := appResource.Render(ctx, input, command)
		if err != nil {
			return nil, err
		}

		for _, unstructuredResource := range renderedResource {
			labels := unstructuredResource.GetLabels()
			if labels == nil {
				labels = make(map[string]string)
			}
			labels[InstallationStepLabel] = a.getLabel(i)
			unstructuredResource.SetLabels(labels)
			allResources = append(allResources, unstructuredResource)
		}
	}
	return allResources, nil
}