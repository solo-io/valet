package application

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Patch struct {
	Path      string `yaml:"path"`
	PatchType string `yaml:"patchType"`
	Name      string `yaml:"name" valet:"template"`
	Namespace string `yaml:"namespace" valet:"template"`
	KubeType  string `yaml:"kubeType"`

	Values render.Values `yaml:"values"`
}

func (p *Patch) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(p.Values)
	if err := input.Values.RenderFields(p); err != nil {
		return err
	}
	cmd.Stdout().Println("Patching %s.%s (%s) from file %s (%s) %s", p.Namespace, p.Name, p.KubeType, p.Path, p.PatchType, input.Values.ToString())
	patchTemplate, err := render.LoadFile(p.Path)
	if err != nil {
		return err
	}
	patchString, err := render.LoadTemplate(patchTemplate, input.Values)
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

func (p *Patch) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(p.Values)
	if err := input.Values.RenderFields(p); err != nil {
		return err
	}
	cmd.Stdout().Println("Skipping teardown for patch")
	return nil
}

func (p *Patch) Load(input render.InputParams) (string, error) {
	t := Template{
		Path:   p.Path,
		Values: input.Values,
	}
	return t.Load(input)
}
