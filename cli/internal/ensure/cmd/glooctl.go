package cmd

type Glooctl Command

func (g *Glooctl) With(args... string) *Glooctl {
	g.Args = append(g.Args, args...)
	return g
}

func (g *Glooctl) Command() Command {
	return Command(*g)
}

func (g *Glooctl) Run() error {
	return g.Command().Run()
}

func (g *Glooctl) Output() (string, error) {
	return g.Command().Output()
}

func (g *Glooctl) UninstallAll() *Glooctl {
	return g.With("uninstall", "--all")
}

func (g *Glooctl) ProxyUrl() *Glooctl {
	return g.With("proxy", "url")
}

func (g *Glooctl) ProxyAddress() *Glooctl {
	return g.With("proxy", "url")
}

func (g *Glooctl) GetUpstream(name string) *Glooctl {
	return g.With("get", "upstream", name)
}

func (g *Glooctl) CreateUpstream(name string) *Glooctl {
	return g.With("create", "upstream", name)
}

func (g *Glooctl) AwsSecretName(secretName string) *Glooctl {
	return g.With("--aws-secret-name", secretName)
}
