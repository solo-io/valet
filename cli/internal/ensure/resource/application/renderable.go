package application

import (
	"context"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"reflect"
)

type Renderable interface {
	Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error)
}

func RenderFirst(ctx context.Context, input render.InputParams, command cmd.Factory, resources ...Renderable) (kuberesource.UnstructuredResources, error) {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		return resource.Render(ctx, input, command)
	}
	return nil, nil
}