package step

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

type Apply struct {
	Path string `json:"path"`
}

func (a *Apply) GetDescription() string {
	stringCmd := a.GetCmd().ToString()
	return fmt.Sprintf("Running command: %s", stringCmd)
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
