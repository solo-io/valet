package resource

import (
	"context"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type ClusterResource interface {
	Resource
	SetContext(ctx context.Context, command cmd.Factory) error
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

func (c *Cluster) SetContext(ctx context.Context, command cmd.Factory) error {
	if c.Minikube != nil {
		return c.Minikube.SetContext(ctx, command)
	}
	if c.GKE != nil {
		return c.GKE.SetContext(ctx, command)
	}
	if c.EKS != nil {
		return c.EKS.SetContext(ctx, command)
	}

	return NoClusterDefinedError
}

func (c *Cluster) Ensure(ctx context.Context, inputs InputParams, command cmd.Factory) error {
	if c.Minikube != nil {
		return c.Minikube.Ensure(ctx, inputs, command)
	}
	if c.GKE != nil {
		return c.GKE.Ensure(ctx, inputs, command)
	}
	if c.EKS != nil {
		return c.EKS.Ensure(ctx, inputs, command)
	}
	return nil
}

func (c *Cluster) Teardown(ctx context.Context, inputs InputParams, command cmd.Factory) error {
	if c.Minikube != nil {
		return c.Minikube.Teardown(ctx, inputs, command)
	}
	if c.GKE != nil {
		return c.GKE.Teardown(ctx, inputs, command)
	}
	if c.EKS != nil {
		return c.EKS.Teardown(ctx, inputs, command)
	}
	return nil
}
