package kubectl

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

var _ api.Step = new(ApplyTemplate)

type ApplyTemplate struct {
	Path string `json:"path,omitempty"`
}

func (a *ApplyTemplate) GetDescription(ctx *api.WorkflowContext, values render.Values) (string, error) {
	command, err := a.GetCmd(ctx, values)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Running command: %s", command.ToString()), nil
}

func (a *ApplyTemplate) GetCmd(ctx *api.WorkflowContext, values render.Values) (*cmd.Command, error) {
	tmpl, err := ctx.FileStore.Load(a.Path)
	if err != nil {
		return nil, err
	}
	loaded, err := render.LoadTemplate(tmpl, values, ctx.Runner)
	if err != nil {
		return nil, err
	}
	return cmd.New().Kubectl().ApplyStdIn(loaded).Cmd(), nil
}

func (a *ApplyTemplate) Run(ctx *api.WorkflowContext, values render.Values) error {
	command, err := a.GetCmd(ctx, values)
	if err != nil {
		return err
	}
	return ctx.Runner.Run(command)
}

func (a *ApplyTemplate) GetDocs(ctx *api.WorkflowContext, values render.Values, flags render.Flags) (string, error) {
	if flags.Contains(DocsFlagYamlOnly) {
		contents, err := ctx.FileStore.Load(a.Path)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("```yaml\n%s\n```", contents), nil
	}
	panic("not implemented")
}
