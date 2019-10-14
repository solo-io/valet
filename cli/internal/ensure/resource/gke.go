package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ ClusterResource = new(GKE)

type GKE struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
	Project  string `yaml:"project"`
}

func (g *GKE) Ensure(ctx context.Context) error {
	gkeClient, err := client.NewGkeClient(ctx)
	if err != nil {
		return err
	}
	running, err := gkeClient.IsRunning(ctx, g.Name, g.Project, g.Location)
	if err != nil {
		return err
	}
	if running {
		return g.SetContext(ctx)
	} else {
		return gkeClient.Create(ctx, g.Name, g.Project, g.Location)
	}
}

func (g *GKE) SetContext(ctx context.Context) error {
	return cmd.Gcloud().GetCredentials().Project(g.Project).Zone(g.Location).WithName(g.Name).Run()
}

func (g *GKE) Teardown(ctx context.Context) error {
	gkeClient, err := client.NewGkeClient(ctx)
	if err != nil {
		return err
	}
	return gkeClient.Destroy(ctx, g.Name, g.Project, g.Location)
}
