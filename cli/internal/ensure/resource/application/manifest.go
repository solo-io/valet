package application

import (
	"context"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

var (
	_ resource.Resource = new(Manifest)
	_ Renderable        = new(Manifest)
)

type Manifest struct {
	Path string `yaml:"path"`
}

func (m *Manifest) Ensure(ctx context.Context, _ render.InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Ensuring manifest %s", m.Path)
	return command.Kubectl().Apply().File(m.Path).Cmd().Run(ctx)
}

func (m *Manifest) Teardown(ctx context.Context, _ render.InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Tearing down manifest %s", m.Path)
	return command.Kubectl().DeleteFile(m.Path).IgnoreNotFound().Cmd().Run(ctx)
}

func (m *Manifest) Render(ctx context.Context, _ render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	cmd.Stdout().Println("Rendering manifest %s", m.Path)
	contents, err := render.LoadFile(m.Path)
	if err != nil {
		return nil, err
	}
	return render.YamlToResources([]byte(contents))
}
