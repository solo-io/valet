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

	Values Values `yaml:"values"`
}

func (p *Patch) updateWithValues(values Values) {
	p.Values = MergeValues(values, p.Values)
}

func (p *Patch) Ensure(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Patching %s.%s (%s) from file %s (%s)", p.Namespace, p.Name, p.KubeType, p.Path, p.PatchType)
	patchString, err := LoadFile(p.Path)
	if err != nil {
		return err
	}
	name, err := LoadTemplate(p.Name, p.Values)
	if err != nil {
		return err
	}
	namespace, err := LoadTemplate(p.Namespace, p.Values)
	if err != nil {
		return err
	}
	kubectl := command.Kubectl().
		With("patch", p.KubeType, name).
		Namespace(namespace).
		With("--type", p.PatchType).
		With("--patch", patchString)
	return kubectl.Cmd().Run(ctx)
}

func (p *Patch) Teardown(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Skipping teardown for patch")
	return nil
}

func (p *Patch) Load() (string, error) {
	t := Template{
		Path:      p.Path,
		Values:    p.Values,
	}
	return t.Load()
}
