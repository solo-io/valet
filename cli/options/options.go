package options

import (
	"context"
)

type Options struct {
	Ensure Ensure
	Config Config
	Build  Build
	Top    Top
}

type Top struct {
	Ctx context.Context
}

type Ensure struct {
	File              string
	ValetArtifacts    bool
	LocalArtifactsDir string
	GkeClusterName    string
	GlooVersion       string
	// if true, attempt teardown after ensure finishes.
	// return error if ensure returned error
	TeardownOnFinish bool

	Values     map[string]string
	Flags      []string
	Step       bool
	DryRun     bool
	Registry   string
	KubeConfig string
}

type Build struct {
	File    string
	Version string
	// if true, then don't push images to docker repo
	SkipDockerPush bool
}

type Config struct {
	GlobalConfigPath string
	RegistryName     string
	RegistryPath     string
}
