package options

import "context"

type Options struct {
	Cluster Cluster
	Gloo    Gloo
	Top     Top
}

type Top struct {
	Ctx context.Context
}

type Cluster struct {
	Type        string
	Name        string
	Location    string
	Project     string
	KubeVersion string
}

type Gloo struct {
	GlooVersion   string
	Enterprise    bool
	LicenseKey    string
	GlooNamespace string
}
