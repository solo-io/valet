package cluster

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

var _ ClusterResource = new(EKS)

type EKS struct {
	Name   string `json:"name"   valet:"template,key=ClusterName"`
	Region string `json:"region" valet:"template,key=AwsRegion,default=us-east-2"`
}

func (e *EKS) Ensure(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(e, ctx.Runner); err != nil {
		return err
	}
	cmd.Stdout().Println("Ensuring eks cluster %s (region : %s)", e.Name, e.Region)
	running, err := cmd.New().EksCtl().IsRunning(e.Name, e.Region, ctx.Runner)
	if err != nil {
		return err
	}
	if running {
		if err := cmd.New().EksCtl().WriteKubeConfig(e.Name, e.Region, ctx.Runner); err != nil {
			return err
		}
		return e.SetContext(ctx, values)
	}
	if err := cmd.New().EksCtl().CreateCluster(e.Name, e.Region, ctx.Runner); err != nil {
		return err
	}
	return cmd.New().EksCtl().WriteKubeConfig(e.Name, e.Region, ctx.Runner)
}

func (e *EKS) Teardown(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(e, ctx.Runner); err != nil {
		return err
	}
	cmd.Stdout().Println("tearing down eks cluster %s (region : %s)", e.Name, e.Region)
	running, err := cmd.New().EksCtl().IsRunning(e.Name, e.Region, ctx.Runner)
	if err != nil {
		return err
	}
	if !running {
		return nil
	}
	return cmd.New().EksCtl().DeleteCluster(e.Name, e.Region, ctx.Runner)
}

func (e *EKS) SetContext(ctx *api.WorkflowContext, values render.Values) error {
	return ctx.Runner.Run(cmd.New().EksCtl().GetCredentials().Region(e.Region).WithName(e.Name).Cmd())
}
