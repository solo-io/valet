package application

import (
	"context"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var (
	_ Renderable = new(Template)
)

type Template struct {
	RegistryName string        `yaml:"registry" valet:"default=default"`
	Path         string        `yaml:"path"`
	Values       render.Values `yaml:"values"`
}

func (t *Template) Load(input render.InputParams) (string, error) {
	input = input.MergeValues(t.Values)
	if err := input.RenderFields(t); err != nil {
		return "", err
	}
	cmd.Stdout().Println("Loading template %s:%s", t.RegistryName, t.Path)
	tmpl, err := input.LoadFile(t.RegistryName, t.Path)
	if err != nil {
		return "", err
	}
	loaded, err := render.LoadTemplate(tmpl, input.Values, input.Runner())
	if err != nil {
		cmd.Stderr().Println("Error loading template: %s", err.Error())
	}
	return loaded, err
}

func (t *Template) Render(ctx context.Context, input render.InputParams) (kuberesource.UnstructuredResources, error) {
	loaded, err := t.Load(input)
	if err != nil {
		return nil, err
	}
	return render.YamlToResources([]byte(loaded))
}
