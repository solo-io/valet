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
	cmd.Stdout().Println("Patching %s.%s (%s) from file %s (%s)", p.Namespace, p.Name, p.KubeType, p.Path, p.PatchType)
	patchString, err := LoadFile(p.Path)
	if err != nil {
		return err
	}
	kubectl := command.Kubectl().
		With("patch", p.KubeType, p.Name).
		Namespace(p.Namespace).
		With("--type", p.PatchType).
		With("--patch", patchString)
	return kubectl.Cmd().Run(ctx)
}

func (p *Patch) Teardown(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Skipping teardown for patch")
	return nil
}
