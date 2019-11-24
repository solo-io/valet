package cluster

import (
	"context"
	"path/filepath"

	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/solo-io/go-utils/errors"
)

type ClusterResource interface {
	resource.Resource
	SetContext(ctx context.Context, input render.InputParams) error
}

var (
	_ ClusterResource = new(Cluster)

	NoClusterDefinedError = errors.Errorf("no cluster defined")
)

type Cluster struct {
	Kind       *Kind     `yaml:"kind"`
	Minikube   *Minikube `yaml:"minikube"`
	GKE        *GKE      `yaml:"gke"`
	EKS        *EKS      `yaml:"eks"`
	KubeConfig string    `yaml:"kubeConfig"valet:"template,key=KubeConfig"`
}

func (c *Cluster) getKubeConfig() error {
	if c.KubeConfig == "" {
		switch {
		case c.Minikube != nil:
			c.KubeConfig = clientcmd.RecommendedHomeFile
		case c.GKE != nil:
			c.KubeConfig = filepath.Join(clientcmd.RecommendedConfigDir, "gke", c.GKE.Name)
		case c.EKS != nil:
			c.KubeConfig = filepath.Join(clientcmd.RecommendedConfigDir, "eksctl", c.EKS.Name)
		case c.Kind != nil:
			c.KubeConfig = filepath.Join(clientcmd.RecommendedConfigDir, "kind", c.Kind.Name)
		}
	}
	return nil
}

func (c *Cluster) SetContext(ctx context.Context, input render.InputParams) error {
	if err := resource.RenderAll(ctx, input, c, c.Minikube, c.GKE, c.EKS, c.Kind); err != nil {
		return err
	}
	if err := c.getKubeConfig(); err != nil {
		return err
	}
	input.SetKubeConfig(c.KubeConfig)
	switch {
	case c.Minikube != nil:
		return c.Minikube.SetContext(ctx, input)
	case c.GKE != nil:
		return c.GKE.SetContext(ctx, input)
	case c.EKS != nil:
		return c.EKS.SetContext(ctx, input)
	case c.Kind != nil:
		return c.EKS.SetContext(ctx, input)
	}

	return NoClusterDefinedError
}

func (c *Cluster) Ensure(ctx context.Context, input render.InputParams) error {
	if err := resource.RenderAll(ctx, input, c, c.Minikube, c.GKE, c.EKS, c.Kind); err != nil {
		return err
	}
	if err := c.getKubeConfig(); err != nil {
		return err
	}
	input.SetKubeConfig(c.KubeConfig)
	switch {
	case c.Minikube != nil:
		return c.Minikube.Ensure(ctx, input)
	case c.GKE != nil:
		return c.GKE.Ensure(ctx, input)
	case c.EKS != nil:
		return c.EKS.Ensure(ctx, input)
	case c.Kind != nil:
		return c.Kind.Ensure(ctx, input)
	}
	return nil
}

func (c *Cluster) Teardown(ctx context.Context, input render.InputParams) error {
	if err := resource.RenderAll(ctx, input, c, c.Minikube, c.GKE, c.EKS, c.Kind); err != nil {
		return err
	}
	if err := c.getKubeConfig(); err != nil {
		return err
	}
	input.SetKubeConfig(c.KubeConfig)
	switch {
	case c.Minikube != nil:
		return c.Minikube.Teardown(ctx, input)
	case c.GKE != nil:
		return c.GKE.Teardown(ctx, input)
	case c.EKS != nil:
		return c.EKS.Teardown(ctx, input)
	case c.Kind != nil:
		return c.Kind.Teardown(ctx, input)
	}
	return nil
}
