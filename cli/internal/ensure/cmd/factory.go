package cmd

type Factory interface {
	SetGlooctlPath(path string)
	Glooctl() *Glooctl
	Kubectl() *kubectl
	Gcloud() *gcloud
	Minikube() *minikube
	Helm() *helm
}

type CommandFactory struct {
	GlooctlLocalPath string
}

func (c *CommandFactory) Glooctl() *Glooctl {
	if c.GlooctlLocalPath == "" {
		return nil
	}
	return &Glooctl{
		Name: c.GlooctlLocalPath,
	}
}

func (c *CommandFactory) SetGlooctlPath(path string) {
	c.GlooctlLocalPath = path
}

func (c *CommandFactory) Kubectl() *kubectl {
	return Kubectl()
}

func (c *CommandFactory) Gcloud() *gcloud {
	return Gcloud()
}

func (c *CommandFactory) Minikube() *minikube {
	return Minikube()
}

func (c *CommandFactory) Helm() *helm {
	return Helm()
}



