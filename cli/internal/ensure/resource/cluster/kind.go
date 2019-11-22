package cluster

import (
	"context"
	"fmt"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ ClusterResource = new(Kind)

type Kind struct {
	 Name        string    `yaml:"name" valet:"template,key=ClusterName,default=test"`
}

func (k *Kind) Ensure(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(k); err != nil {
		return err
	}
	cmd.Stdout().Println("Ensuring kind cluster (name: %s)", k.Name)
	running, err := cmd.New().Kind().IsRunning(ctx, input.Runner(), k.Name)
	if err != nil {
		return err
	}
	if running {
		return k.SetContext(ctx, input.Runner())
	}
	return cmd.New().Kind().CreateCluster(ctx, input.Runner(), k.Name)
}

func (k *Kind) SetContext(ctx context.Context, runner cmd.Runner) error {
	return runner.Run(ctx, cmd.New().Kubectl().UseContext(fmt.Sprintf("kind-%s", k.Name)).Cmd())
}

func (k *Kind) Teardown(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(k); err != nil {
		return err
	}
	cmd.Stdout().Println("Tearing down kind cluster")
	return cmd.New().Kind().DeleteCluster(ctx, input.Runner(), k.Name)
}
