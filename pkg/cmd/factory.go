package cmd

const (
	KubectlCmd  = "kubectl"
	HelmCmd     = "helm"
	GcloudCmd   = "gcloud"
	MinikubeCmd = "minikube"
	EksCtlCmd   = "eksctl"
)

type Factory interface {
	SetLocalPath(path, localPath string)
	Kubectl() *Kubectl
	Gcloud() *Gcloud
	Minikube() *Minikube
	EksCtl() *EksCtl
}

func New() Factory {
	return &CommandFactory{}
}

type CommandFactory struct {
	LocalPathOverride map[string]string
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
		Name: c.getLocalPath(path),
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
		cmd: c.getCommand(KubectlCmd),
	}
}

func (c *CommandFactory) Gcloud() *Gcloud {
	return &Gcloud{
		cmd: c.getCommand(GcloudCmd),
	}
}

func (c *CommandFactory) Minikube() *Minikube {
	return &Minikube{
		cmd: c.getCommand(MinikubeCmd),
	}
}

func (c *CommandFactory) EksCtl() *EksCtl {
	return &EksCtl{
		cmd: c.getCommand(EksCtlCmd),
	}
}
