package resource

import (
	"context"
	"github.com/solo-io/go-utils/errors"
)

type ClusterResource interface {
	Resource
	SetContext(ctx context.Context) error
}

var (
	_ ClusterResource = new(Cluster)

	NoClusterDefinedError = errors.Errorf("no cluster defined")
)

type Cluster struct {
	Minikube *Minikube `yaml:"minikube"`
	GKE      *GKE      `yaml:"gke"`
}

func (c *Cluster) SetContext(ctx context.Context) error {
	if c.Minikube != nil {
		return c.Minikube.SetContext(ctx)
	}
	if c.GKE != nil {
		return c.GKE.SetContext(ctx)
	}
	return NoClusterDefinedError
}

func (c *Cluster) Ensure(ctx context.Context) error {
	if c.Minikube != nil {
		return c.Minikube.Ensure(ctx)
	}
	if c.GKE != nil {
		return c.GKE.Ensure(ctx)
	}
	return nil
}

func (c *Cluster) Teardown(ctx context.Context) error {
	if c.Minikube != nil {
		return c.Minikube.Teardown(ctx)
	}
	if c.GKE != nil {
		return c.GKE.Teardown(ctx)
	}
	return nil
}
