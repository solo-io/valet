package cluster

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ ClusterResource = new(EKS)



type EKS struct {
	Name   string `yaml:"name"   valet:"template,key=ClusterName"`
	Region string `yaml:"region" valet:"template,key=AwsRegion,default=us-east-2"`
}

func (e *EKS) Ensure(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(e); err != nil {
		return err
	}
	cmd.Stdout(ctx).Println("Ensuring eks cluster %s (region : %s)", e.Name, e.Region)
	running, err := cmd.New().EksCtl().IsRunning(ctx, e.Name, e.Region, input.Runner())
	if err != nil {
		return err
	}
	if running {
		if err := cmd.New().EksCtl().WriteKubeConfig(ctx, e.Name, e.Region, input.Runner()); err != nil {
			return err
		}
		return e.SetContext(ctx, input)
	}
	if err := cmd.New().EksCtl().KubeConfig(input.KubeConfig()).
		CreateCluster(ctx, e.Name, e.Region, input.Runner()); err != nil {
		return err
	}
	return cmd.New().EksCtl().WriteKubeConfig(ctx, e.Name, e.Region, input.Runner())
}

func (e *EKS) Teardown(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(e); err != nil {
		return err
	}
	cmd.Stdout(ctx).Println("tearing down eks cluster %s (region : %s)", e.Name, e.Region)
	running, err := cmd.New().EksCtl().KubeConfig(input.KubeConfig()).IsRunning(ctx, e.Name, e.Region, input.Runner())
	if err != nil {
		return err
	}
	if !running {
		return nil
	}
	return cmd.New().EksCtl().DeleteCluster(ctx, e.Name, e.Region, input.Runner())
}

func (e *EKS) SetContext(ctx context.Context, input render.InputParams) error {
	return input.Runner().Run(ctx, cmd.New().EksCtl().KubeConfig(input.KubeConfig()).
		GetCredentials().Region(e.Region).WithName(e.Name).Cmd())
}
