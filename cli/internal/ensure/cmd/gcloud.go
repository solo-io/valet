package cmd

import "fmt"

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

func (g *gcloud) Command() Command {
	return Command(*g)
}

func (g *gcloud) Run() error {
	return g.Command().Run()
}

func (g *gcloud) Output() (string, error) {
	return g.Command().Output()
}

func Gcloud(args ...string) *gcloud {
	return &gcloud{
		Name: "gcloud",
		Args: args,
	}
}