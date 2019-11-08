package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ ClusterResource = new(EKS)

type EKS struct {
	Name    string                `yaml:"name"`
	Region  string                `yaml:"region"`
}

func (e *EKS) Ensure(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Ensuring eks cluster %s (region : %s)", e.Name, e.Region)
	running, err := command.EksCtl().IsRunning(ctx, e.Name, e.Region)
	if err != nil {
		return err
	}
	if running {
		return nil
	}
	return command.EksCtl().CreateCluster(ctx, e.Name, e.Region)
}

func (e *EKS) SetContext(ctx context.Context, command cmd.Factory) error {
	return command.EksCtl().GetCredentials().Region(e.Region).WithName(e.Name).Cmd().Run(ctx)
}

func (e *EKS) Teardown(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("tearing down eks cluster %s (region : %s)", e.Name, e.Region)
	running, err := command.EksCtl().IsRunning(ctx, e.Name, e.Region)
	if err != nil {
		return err
	}
	if !running {
		return nil
	}
	return command.EksCtl().DeleteCluster(ctx, e.Name, e.Region)
}
