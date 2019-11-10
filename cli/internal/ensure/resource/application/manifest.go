package application

import (
	"context"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var (
	_ Renderable = new(Manifest)
)

type Manifest struct {
	RegistryName string `yaml:"registry" valet:"default=default"`
	Path         string `yaml:"path"`
}

func (m *Manifest) Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	cmd.Stdout().Println("Rendering manifest %s", m.Path)
	contents, err := m.load(input)
	if err != nil {
		return nil, err
	}
	return render.YamlToResources([]byte(contents))
}

func (m *Manifest) load(input render.InputParams) (string, error) {
	if err := input.Values.RenderFields(m); err != nil {
		return "", err
	}
	contents, err := input.LoadFile(m.RegistryName, m.Path)
	if err != nil {
		return "", err
	}
	return contents, nil
}
