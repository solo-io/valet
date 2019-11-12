package application

import (
	"context"
	"reflect"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

type Renderable interface {
	Render(ctx context.Context, input render.InputParams) (kuberesource.UnstructuredResources, error)
}

func RenderFirst(ctx context.Context, input render.InputParams, resources ...Renderable) (kuberesource.UnstructuredResources, error) {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		return resource.Render(ctx, input)
	}
	return nil, nil
}

