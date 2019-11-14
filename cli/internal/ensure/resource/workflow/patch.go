package workflow

import (
	"context"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var (
	UnableToLoadPatchError = func(err error) error {
		return errors.Wrapf(err, "unable to load patch")
	}
)

type Patch struct {
	RegistryName string `yaml:"registry" valet:"default=default"`
	Path         string `yaml:"path"`
	PatchType    string `yaml:"patchType"`
	Name         string `yaml:"name" valet:"template"`
	Namespace    string `yaml:"namespace" valet:"template"`
	KubeType     string `yaml:"kubeType"`

	Values render.Values `yaml:"values"`
}

func (p *Patch) Ensure(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(p.Values)
	if err := input.RenderFields(p); err != nil {
		return err
	}
	cmd.Stdout().Println("Patching %s.%s (%s) from file %s (%s) %s", p.Namespace, p.Name, p.KubeType, p.Path, p.PatchType, input.Values.ToString())
	patchTemplate, err := input.LoadFile(p.RegistryName, p.Path)
	if err != nil {
		return UnableToLoadPatchError(err)
	}
	patchString, err := render.LoadTemplate(patchTemplate, input.Values, input.Runner())
	if err != nil {
		return err
	}
	kubectl := cmd.New().Kubectl().
		With("patch", p.KubeType, p.Name).
		Namespace(p.Namespace).
		With("--type", p.PatchType).
		With("--patch", patchString).
		Cmd()
	return input.Runner().Run(ctx, kubectl)
}

func (p *Patch) Teardown(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(p.Values)
	if err := input.RenderFields(p); err != nil {
		return err
	}
	cmd.Stdout().Println("Skipping teardown for patch")
	return nil
}
