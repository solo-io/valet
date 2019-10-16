package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	PetstoreDemoName = "petstore"
)

var (
	petstoreFiles = []string{
		"https://raw.githubusercontent.com/sololabs/demos/b523571c66057a5591bce22ad896729f1fee662b/petclinic_demo/petstore.yaml",
	}
	petstoreResources = Resources{Paths: petstoreFiles}
)

type Petstore struct {
}

func (p *Petstore) Ensure(ctx context.Context, command cmd.Factory) error {
	return petstoreResources.Ensure(ctx, command)
}

func (p *Petstore) Teardown(ctx context.Context, command cmd.Factory) error {
	return petstoreResources.Teardown(ctx, command)
}
