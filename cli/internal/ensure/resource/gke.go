package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ ClusterResource = new(GKE)

type GKE struct {
	Name     string                `yaml:"name"`
	Location string                `yaml:"location"`
	Project  string                `yaml:"project"`
	Options  *client.CreateOptions `yaml:"options"`
}

func (g *GKE) Ensure(ctx context.Context, command cmd.Factory) error {
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
	return g.SetContext(ctx, command)
}

func (g *GKE) SetContext(ctx context.Context, command cmd.Factory) error {
	return command.Gcloud().GetCredentials().Project(g.Project).Zone(g.Location).WithName(g.Name).Cmd().Run(ctx)
}

func (g *GKE) Teardown(ctx context.Context, command cmd.Factory) error {
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
