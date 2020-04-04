package kubectl

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

var _ api.Step = new(Apply)

type Apply struct {
	Path string `json:"path"`
}

func (a *Apply) GetDescription(_ *api.WorkflowContext, _ render.Values) (string, error) {
	stringCmd := a.GetCmd().ToString()
	return fmt.Sprintf("Running command: %s", stringCmd), nil
}

func (a *Apply) GetCmd() *cmd.Command {
	return cmd.New().Kubectl().ApplyFile(a.Path).Cmd()
}

func (a *Apply) Run(ctx *api.WorkflowContext, values render.Values) error {
	return ctx.Runner.Run(a.GetCmd())
}

func (a *Apply) GetDocs(ctx *api.WorkflowContext, options api.DocsOptions) (string, error) {
	panic("implement me")
}
