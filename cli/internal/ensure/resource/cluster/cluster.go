package cluster

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type ClusterResource interface {
	resource.Resource
	SetContext(ctx context.Context, runner cmd.Runner) error
}

var (
	_ ClusterResource = new(Cluster)

	NoClusterDefinedError = errors.Errorf("no cluster defined")
)

type Cluster struct {
	Kind     *Kind     `yaml:"kind"`
	Minikube *Minikube `yaml:"minikube"`
	GKE      *GKE      `yaml:"gke"`
	EKS      *EKS      `yaml:"eks"`
}

func (c *Cluster) SetContext(ctx context.Context, runner cmd.Runner) error {
	switch {
	case c.Minikube != nil:
		return c.Minikube.SetContext(ctx, runner)
	case c.GKE != nil:
		return c.GKE.SetContext(ctx, runner)
	case c.EKS != nil:
		return c.EKS.SetContext(ctx, runner)
	case c.Kind != nil:
		return c.EKS.SetContext(ctx, runner)
	}

	return NoClusterDefinedError
}

func (c *Cluster) Ensure(ctx context.Context, inputs render.InputParams) error {
	switch {
	case c.Minikube != nil:
		return c.Minikube.Ensure(ctx, inputs)
	case c.GKE != nil:
		return c.GKE.Ensure(ctx, inputs)
	case c.EKS != nil:
		return c.EKS.Ensure(ctx, inputs)
	case c.Kind != nil:
		return c.Kind.Ensure(ctx, inputs)
	}
	return nil
}

func (c *Cluster) Teardown(ctx context.Context, inputs render.InputParams) error {
	switch {
	case c.Minikube != nil:
		return c.Minikube.Teardown(ctx, inputs)
	case c.GKE != nil:
		return c.GKE.Teardown(ctx, inputs)
	case c.EKS != nil:
		return c.EKS.Teardown(ctx, inputs)
	case c.Kind != nil:
		return c.Kind.Teardown(ctx, inputs)
	}
	return nil
}
