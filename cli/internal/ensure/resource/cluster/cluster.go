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
	Minikube *Minikube `yaml:"minikube"`
	GKE      *GKE      `yaml:"gke"`
	EKS      *EKS      `yaml:"eks"`
}

func (c *Cluster) SetContext(ctx context.Context, runner cmd.Runner) error {
	if c.Minikube != nil {
		return c.Minikube.SetContext(ctx, runner)
	}
	if c.GKE != nil {
		return c.GKE.SetContext(ctx, runner)
	}
	if c.EKS != nil {
		return c.EKS.SetContext(ctx, runner)
	}

	return NoClusterDefinedError
}

func (c *Cluster) Ensure(ctx context.Context, inputs render.InputParams) error {
	if c.Minikube != nil {
		return c.Minikube.Ensure(ctx, inputs)
	}
	if c.GKE != nil {
		c.GKE.SetValues(inputs)
		return c.GKE.Ensure(ctx, inputs)
	}
	if c.EKS != nil {
		c.EKS.SetValues(inputs)
		return c.EKS.Ensure(ctx, inputs)
	}
	return nil
}

func (c *Cluster) Teardown(ctx context.Context, inputs render.InputParams) error {
	if c.Minikube != nil {
		return c.Minikube.Teardown(ctx, inputs)
	}
	if c.GKE != nil {
		c.GKE.SetValues(inputs)
		return c.GKE.Teardown(ctx, inputs)
	}
	if c.EKS != nil {
		c.EKS.SetValues(inputs)
		return c.EKS.Teardown(ctx, inputs)
	}
	return nil
}
