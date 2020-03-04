package application

import (
	"context"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

// For when you really just want to deploy a directory full of yaml concisely.
type Manifests struct {
	RegistryName string   `json:"registry" valet:"default=default"`
	Paths        []string `json:"paths"`
}

func (m *Manifests) Render(ctx context.Context, input render.InputParams) (kuberesource.UnstructuredResources, error) {
	var resources kuberesource.UnstructuredResources
	for _, path := range m.Paths {
		manifest := Manifest{
			RegistryName: m.RegistryName,
			Path:         path,
		}
		r, err := manifest.Render(ctx, input)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r...)
	}
	return resources, nil
}
