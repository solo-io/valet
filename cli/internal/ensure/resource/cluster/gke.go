package cluster

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ ClusterResource = new(GKE)

const (
	ProjectKey  = "Project"
	LocationKey = "Location"
)

type GKE struct {
	Name     string                `yaml:"name"`
	Location string                `yaml:"location"`
	Project  string                `yaml:"project"`
	Options  *client.CreateOptions `yaml:"options"`
}

func (g *GKE) Ensure(ctx context.Context, input render.InputParams) error {
	cmd.Stdout().Println("Ensuring GKE cluster %s (project: %s, location: %s)", g.Name, g.Project, g.Location)
	gkeClient, err := client.NewGkeClient(ctx)
	if err != nil {
		return err
	}
	running, err := gkeClient.IsRunning(ctx, g.Name, g.Project, g.Location)
	if err != nil {
		return err
	}
	if !running {
		if err := gkeClient.Create(ctx, g.Name, g.Project, g.Location, g.Options); err != nil {
			return err
		}
	}
	return g.SetContext(ctx, input.Runner())
}

func (g *GKE) SetContext(ctx context.Context, runner cmd.Runner) error {
	return runner.Run(ctx, cmd.New().Gcloud().GetCredentials().Project(g.Project).Zone(g.Location).WithName(g.Name).Cmd())
}

func (g *GKE) SetValues(input render.InputParams) {
	input.Values[render.ClusterKey] = g.Name
	input.Values[ProjectKey] = g.Project
	input.Values[LocationKey] = g.Location
}

func (g *GKE) Teardown(ctx context.Context, input render.InputParams) error {
	gkeClient, err := client.NewGkeClient(ctx)
	if err != nil {
		return err
	}
	running, err := gkeClient.IsRunning(ctx, g.Name, g.Project, g.Location)
	if err != nil {
		return err
	} else if !running {
		return nil
	}
	return gkeClient.Destroy(ctx, g.Name, g.Project, g.Location)
}
