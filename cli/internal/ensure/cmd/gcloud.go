package cmd

import (
	"context"
	"fmt"
)

type gcloud Command

func (g *gcloud) GetCredentials() *gcloud {
	return g.With("container", "clusters", "get-credentials")
}

func (g *gcloud) Project(project string) *gcloud {
	return g.With(fmt.Sprintf("--project=%s", project))
}

func (g *gcloud) Zone(zone string) *gcloud {
	return g.With(fmt.Sprintf("--zone=%s", zone))
}

func (g *gcloud) WithName(name string) *gcloud {
	return g.With(name)
}

func (g *gcloud) With(args ...string) *gcloud {
	g.Args = append(g.Args, args...)
	return g
}

func (g *gcloud) WithStdIn(stdIn string) *gcloud {
	g.StdIn = stdIn
	return g
}

func (g *gcloud) Command() *Command {
	return &Command{
		Name:  g.Name,
		Args:  g.Args,
		StdIn: g.StdIn,
	}
}

func (g *gcloud) Run(ctx context.Context) error {
	return g.Command().Run(ctx)
}

func (g *gcloud) Output(ctx context.Context) (string, error) {
	return g.Command().Output(ctx)
}

func Gcloud(args ...string) *gcloud {
	return &gcloud{
		Name: "gcloud",
		Args: args,
	}
}
