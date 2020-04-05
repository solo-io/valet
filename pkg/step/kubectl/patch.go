package kubectl

import (
	"fmt"
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

var _ api.Step = new(Patch)

var (
	UnableToLoadPatchError = func(err error) error {
		return errors.Wrapf(err, "unable to load patch")
	}
)

type Patch struct {
	Path         string `json:"path"`
	PatchType    string `json:"patchType"`
	Name         string `json:"name" valet:"template"`
	Namespace    string `json:"namespace" valet:"template"`
	KubeType     string `json:"kubeType"`
}

func (p *Patch) GetCmd(ctx *api.WorkflowContext, values render.Values) (*cmd.Command, error) {
	if err := values.RenderFields(p, ctx.Runner); err != nil {
		return nil, err
	}
	patchTemplate, err := ctx.FileStore.Load(p.Path)
	if err != nil {
		return nil, UnableToLoadPatchError(err)
	}
	patchString, err := render.LoadTemplate(patchTemplate, values, ctx.Runner)
	if err != nil {
		return nil, err
	}
	kubectl := cmd.New().Kubectl().
		With("patch", p.KubeType, p.Name).
		Namespace(p.Namespace).
		With("--type", p.PatchType).
		With("--patch", patchString).
		Cmd()
	return kubectl, nil
}

func (p *Patch) GetDescription(ctx *api.WorkflowContext, values render.Values) (string, error) {
	kubectl, err := p.GetCmd(ctx, values)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Running command: %s", kubectl.ToString()), nil
}

func (p *Patch) Run(ctx *api.WorkflowContext, values render.Values) error {
	kubectl, err := p.GetCmd(ctx, values)
	if err != nil {
		return err
	}
	return ctx.Runner.Run(kubectl)
}

func (p *Patch) GetDocs(ctx *api.WorkflowContext, values render.Values, flags render.Flags) (string, error) {
	panic("implement me")
}

