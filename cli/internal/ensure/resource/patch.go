package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Patch struct {
	Path      string `yaml:"path"`
	PatchType string `yaml:"patchType"`
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	KubeType  string `yaml:"kubeType"`
}

func (p *Patch) Ensure(ctx context.Context, command cmd.Factory) error {
	patchString, err := LoadFile(ctx, p.Path)
	if err != nil {
		return err
	}
	cmd := command.Kubectl().
		With("patch", p.KubeType, p.Name).
		Namespace(p.Namespace).
		With("--type", p.PatchType).
		With("--patch", patchString)
	return cmd.Cmd().Run(ctx)
}

func (p *Patch) Teardown(ctx context.Context, command cmd.Factory) error {
	return nil
}
