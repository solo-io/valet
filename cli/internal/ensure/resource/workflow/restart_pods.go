package workflow

import (
	"context"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

type RestartPods struct {
	Namespace string   `yaml:"namespace" valet:"key=Namespace"`
	Labels    []string `yaml:"labels"`
}

func (r *RestartPods) Ensure(ctx context.Context, inputs render.InputParams) error {
	cmd.Stdout().Println("Restarting pods for in namespace %s", r.Namespace)
	command := cmd.New().Kubectl().Delete("pod").Namespace(r.Namespace)
	if len(r.Labels) == 0 {
		command = command.With("--all")
	} else {
		for _, label := range r.Labels {
			command = command.With("-l").With(label)
		}
	}
	err := inputs.Runner().Run(ctx, command.Cmd())
	if err != nil {
		return err
	}
	err = internal.WaitUntilPodsRunning(r.Namespace)
	if err != nil {
		return err
	}
	cmd.Stdout().Println("Done restarting pods")
	return nil
}

func (r *RestartPods) Teardown(ctx context.Context, inputs render.InputParams) error {
	return nil
}
