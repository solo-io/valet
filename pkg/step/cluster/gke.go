package cluster

import (
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/client/gke"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

var _ ClusterResource = new(GKE)

type GKE struct {
	Name     string            `json:"name" valet:"template,key=ClusterName"`
	Location string            `json:"location" valet:"template,key=GcloudLocation"`
	Project  string            `json:"project" valet:"template,key=GcloudProject"`
	Options  gke.CreateOptions `json:"options"`
}

func (g *GKE) Ensure(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(g, ctx.Runner); err != nil {
		return err
	}
	cmd.Stdout().Println("Ensuring GKE cluster %s (project: %s, location: %s)", g.Name, g.Project, g.Location)
	gkeClient, err := gke.NewClient(ctx.Ctx)
	if err != nil {
		return err
	}
	running, err := gkeClient.IsRunning(ctx.Ctx, g.Name, g.Project, g.Location)
	if err != nil {
		return err
	}
	if !running {
		if err := gkeClient.Create(ctx.Ctx, g.Name, g.Project, g.Location, &g.Options); err != nil {
			return err
		}
	}
	return g.SetContext(ctx, values)
}

func (g *GKE) SetContext(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(g, ctx.Runner); err != nil {
		return err
	}
	return ctx.Runner.Run(cmd.New().Gcloud().GetCredentials().Project(g.Project).Zone(g.Location).WithName(g.Name).Cmd())
}

func (g *GKE) Teardown(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(g, ctx.Runner); err != nil {
		return err
	}
	gkeClient, err := gke.NewClient(ctx.Ctx)
	if err != nil {
		return err
	}
	running, err := gkeClient.IsRunning(ctx.Ctx, g.Name, g.Project, g.Location)
	if err != nil {
		return err
	} else if !running {
		return nil
	}
	return gkeClient.Destroy(ctx.Ctx, g.Name, g.Project, g.Location)
}
