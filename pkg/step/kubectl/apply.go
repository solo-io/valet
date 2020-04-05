package kubectl

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

const (
	DocsFlagYamlOnly = "YamlOnly"
)

var _ api.Step = new(Apply)

type Apply struct {
	Path string `json:"path,omitempty"`
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

func (a *Apply) GetDocs(ctx *api.WorkflowContext, values render.Values, flags render.Flags) (string, error) {
	if flags.Contains(DocsFlagYamlOnly) {
		contents, err := ctx.FileStore.Load(a.Path)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("```yaml\n%s\n```", contents), nil
	}
	return fmt.Sprintf("```\nkubectl apply -f %s\n```", a.Path), nil
}
