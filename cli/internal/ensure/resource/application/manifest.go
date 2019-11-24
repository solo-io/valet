package application

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var (
	_ Renderable = new(Manifest)
)

type Manifest struct {
	RegistryName string `yaml:"registry" valet:"default=default"`
	Path         string `yaml:"path" valet:"key=Path"`
}

func (m *Manifest) Render(ctx context.Context, input render.InputParams) (kuberesource.UnstructuredResources, error) {
	contents, err := m.load(ctx, input)
	if err != nil {
		return nil, err
	}
	return render.YamlToResources([]byte(contents))
}

func (m *Manifest) load(ctx context.Context, input render.InputParams) (string, error) {
	if err := input.RenderFields(m); err != nil {
		return "", err
	}
	manifest := m.Path
	if m.RegistryName != "" && m.RegistryName != render.DefaultRegistry {
		manifest = fmt.Sprintf("%s:%s", m.RegistryName, manifest)
	}
	cmd.Stdout(ctx).Printf("Loading manifest %s", manifest)
	contents, err := input.LoadFile(ctx, m.RegistryName, m.Path)
	if err != nil {
		return "", err
	}
	return contents, nil
}
