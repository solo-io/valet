package options

import "context"

type Options struct {
	Cluster Cluster
	Gloo    Gloo
	Ensure  Ensure
	Top     Top
}

type Top struct {
	Ctx context.Context
}

type Ensure struct {
	File string
}

type Cluster struct {
	Minikube    Minikube `yaml:"minikube"`
	GKE         GKE      `yaml:"gke"`
	Type        string   `yaml:"type"`
	KubeVersion string   `yaml:"kube_version"`
}

type GKE struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
	Project  string `yaml:"project"`
}

type Minikube struct{}

type Gloo struct {
	Version    string `yaml:"version"`
	Enterprise bool   `yaml:"enterprise"`
	LicenseKey string `yaml:"license_key"`
	Namespace  string `yaml:"namespace"`
	AWS        AWS    `yaml:"aws"`
}

type AWS struct {
	Secret   bool `yaml:"secret"`
	Upstream bool `yaml:"upstream"`
}

type Demos struct {
	Petclinic *Petclinic `yaml:"petclinic"`
}

type Petclinic struct{}
