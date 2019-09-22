package options

import (
	"context"
)

type Options struct {
	Ensure Ensure
	Build  Build
	Top    Top
}

type Top struct {
	Ctx                context.Context
}

type Ensure struct {
	File      string
	ValetArtifacts bool
	LocalArtifactsDir string
	GkeClusterName string
	GlooVersion string
}

type Build struct {
	File    string
	Version string
	// if true, then don't push images to docker repo
	SkipDockerPush bool
}
