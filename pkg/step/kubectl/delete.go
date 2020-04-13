package kubectl

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

var _ api.Step = new(Delete)

type Delete struct {
	Path string `json:"path,omitempty"`
}

func (a *Delete) GetDescription(_ *api.WorkflowContext, _ render.Values) (string, error) {
	stringCmd := a.GetCmd().ToString()
	return fmt.Sprintf("Running command: %s", stringCmd), nil
}

func (a *Delete) GetCmd() *cmd.Command {
	return cmd.New().Kubectl().DeleteFile(a.Path).Cmd()
}

func (a *Delete) Run(ctx *api.WorkflowContext, values render.Values) error {
	return ctx.Runner.Run(a.GetCmd())
}

func (a *Delete) GetDocs(ctx *api.WorkflowContext, values render.Values, flags render.Flags) (string, error) {
	if flags.Contains(DocsFlagYamlOnly) {
		contents, err := ctx.FileStore.Load(a.Path)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("```yaml\n%s\n```", contents), nil
	}
	return fmt.Sprintf("```\nkubectl delete -f %s\n```", a.Path), nil
}
