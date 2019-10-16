package cmd

import (
	"fmt"
)

type Gcloud struct {
	cmd *Command
}

func (g *Gcloud) Cmd() *Command {
	return g.cmd
}

func (g *Gcloud) With(args... string) *Gcloud {
	g.cmd = g.cmd.With(args...)
	return g
}

func (g *Gcloud) GetCredentials() *Gcloud {
	return g.With("container", "clusters", "get-credentials")
}

func (g *Gcloud) Project(project string) *Gcloud {
	return g.With(fmt.Sprintf("--project=%s", project))
}

func (g *Gcloud) Zone(zone string) *Gcloud {
	return g.With(fmt.Sprintf("--zone=%s", zone))
}

func (g *Gcloud) WithName(name string) *Gcloud {
	return g.With(name)
}