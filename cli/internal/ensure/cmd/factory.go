package cmd

const (
	GlooctlCmd = "glooctl"
)

type Factory interface {
	SetLocalPath(path, localPath string)
	Glooctl() *Glooctl
	Kubectl() *Kubectl
	Gcloud() *Gcloud
	Minikube() *Minikube
	Helm() *Helm
}

type CommandFactory struct {
	LocalPathOverride map[string]string
	CommandRunner     CommandRunner
}

func (c *CommandFactory) getLocalPath(path string) string {
	if c.LocalPathOverride == nil {
		return path
	}
	if localPath, ok := c.LocalPathOverride[path]; ok {
		return localPath
	}
	return path
}

func (c *CommandFactory) getCommand(path string) *Command {
	return &Command{
		Name: path,
		CommandRunner: c.CommandRunner,
	}
}

func (c *CommandFactory) Glooctl() *Glooctl {
	return &Glooctl{
		cmd: c.getCommand(GlooctlCmd),
	}
}

func (c *CommandFactory) SetLocalPath(path, localPath string) {
	if c.LocalPathOverride == nil {
		c.LocalPathOverride = make(map[string]string)
	}
	c.LocalPathOverride[path] = localPath
}

func (c *CommandFactory) Kubectl() *Kubectl {
	return &Kubectl{
		cmd: c.getCommand("kubectl"),
	}
}

func (c *CommandFactory) Gcloud() *Gcloud {
	return &Gcloud{
		cmd: c.getCommand("gcloud"),
	}
}

func (c *CommandFactory) Minikube() *Minikube {
	return &Minikube{
		cmd: c.getCommand("minikube"),
	}
}

func (c *CommandFactory) Helm() *Helm {
	return &Helm{
		cmd: c.getCommand("helm"),
	}
}
