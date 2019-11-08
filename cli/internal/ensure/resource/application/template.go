package application

import (
	"context"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var (
	_ resource.Resource = new(Template)
	_ Renderable        = new(Template)
)

type Template struct {
	Path   string        `yaml:"path"`
	Values render.Values `yaml:"values"`
}

func (t *Template) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(t.Values)
	cmd.Stdout().Println("Ensuring template %s %s", t.Path, input.Values.ToString())
	rendered, err := t.Load(input)
	if err != nil {
		return err
	}
	return command.Kubectl().ApplyStdIn(rendered).Cmd().Run(ctx)
}

func (t *Template) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(t.Values)
	cmd.Stdout().Println("Tearing down template %s %s", t.Path, input.Values.ToString())
	rendered, err := t.Load(input)
	if err != nil {
		return err
	}
	return command.Kubectl().DeleteStdIn(rendered).Cmd().Run(ctx)
}

func (t *Template) Load(input render.InputParams) (string, error) {
	tmpl, err := render.LoadFile(t.Path)
	if err != nil {
		return "", err
	}
	return render.LoadTemplate(tmpl, input.Values)
}

func (t *Template) Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	input = input.MergeValues(t.Values)
	contents, err := render.LoadFile(t.Path)
	if err != nil {
		return nil, err
	}
	return render.YamlToResources([]byte(contents))
}