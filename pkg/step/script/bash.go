package script

import (
	"fmt"
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

type Bash struct {
	Inline string `json:"inline,omitempty"`
	Path   string `json:"path,omitempty"`
}

func (b *Bash) GetDescription(ctx *api.WorkflowContext, values render.Values) (string, error) {
	if b.Inline != "" {
		return fmt.Sprintf("Running command: bash -c '%s'", b.Inline), nil
	} else if b.Path != "" {
		return fmt.Sprintf("Running command: bash %s", b.Path), nil
	}
	return "", errors.Errorf("Unknown script")
}

func (b *Bash) Run(ctx *api.WorkflowContext, values render.Values) error {
	var args []string
	if b.Inline != "" {
		args = append(args, "-c", b.Inline)
	} else if b.Path != "" {
		args = append(args, b.Path)
	}
	command := cmd.Command{
		Name:            "bash",
		Args:            args,
	}
	return ctx.Runner.Run(&command)
}

func (b *Bash) GetDocs(ctx *api.WorkflowContext, values render.Values, flags render.Flags) (string, error) {
	panic("implement me")
}

