package cluster

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ ClusterResource = new(EKS)

type EKS struct {
	Name   string `yaml:"name"`
	Region string `yaml:"region"`
}

func (e *EKS) Ensure(ctx context.Context, input render.InputParams) error {
	cmd.Stdout().Println("Ensuring eks cluster %s (region : %s)", e.Name, e.Region)
	running, err := cmd.New().EksCtl().IsRunning(ctx, e.Name, e.Region, input.Runner())
	if err != nil {
		return err
	}
	if running {
		return e.SetContext(ctx, input.Runner())
	}
	return cmd.New().EksCtl().CreateCluster(ctx, e.Name, e.Region, input.Runner())
}

func (e *EKS) Teardown(ctx context.Context, input render.InputParams) error {
	cmd.Stdout().Println("tearing down eks cluster %s (region : %s)", e.Name, e.Region)
	running, err := cmd.New().EksCtl().IsRunning(ctx, e.Name, e.Region, input.Runner())
	if err != nil {
		return err
	}
	if !running {
		return nil
	}
	return cmd.New().EksCtl().DeleteCluster(ctx, e.Name, e.Region, input.Runner())
}

func (e *EKS) SetContext(ctx context.Context, runner cmd.Runner) error {
	return runner.Run(ctx, cmd.New().EksCtl().GetCredentials().Region(e.Region).WithName(e.Name).Cmd())
}
