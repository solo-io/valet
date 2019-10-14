package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Config struct {
	Cluster        *Cluster        `yaml:"cluster"`
	Namespaces     []Namespace     `yaml:"namespaces"`
	Secrets        []Secret        `yaml:"secrets"`
	CertManager    *CertManager    `yaml:"certManager"`
	Gloo           *Gloo           `yaml:"gloo"`
	ServiceMeshHub *ServiceMeshHub `yaml:"serviceMeshHub"`
	Workflows      []Workflow      `yaml:"workflows"`
	Demos          *Demos          `yaml:"demos"`
	Resources      []string        `yaml:"resources"`
}

func (c *Config) Ensure(ctx context.Context) error {
	if c.Cluster != nil {
		if err := c.Cluster.Ensure(ctx); err != nil {
			return err
		}
	}
	for _, ns := range c.Namespaces {
		if err := ns.Ensure(ctx); err != nil {
			return err
		}
	}
	for _, secret := range c.Secrets {
		if err := secret.Ensure(ctx); err != nil {
			return err
		}
	}
	if c.CertManager != nil {
		if err := c.CertManager.Ensure(ctx); err != nil {
			return err
		}
	}
	if c.Gloo != nil {
		if err := c.Gloo.Ensure(ctx); err != nil {
			return err
		}
	}
	if c.ServiceMeshHub != nil {
		if err := c.ServiceMeshHub.Ensure(ctx); err != nil {
			return err
		}
	}
	proxyAddress, err := c.Gloo.GetProxyAddress(ctx)
	if err != nil {
		return err
	}
	for _, workflow := range c.Workflows {
		workflow.URL = proxyAddress
		if err := workflow.Ensure(ctx); err != nil {
			return err
		}
	}
	if c.Demos != nil {
		c.Demos.Glooctl = c.Gloo.glooctl
		if err := c.Demos.Ensure(ctx); err != nil {
			return err
		}
	}
	for _, resource := range c.Resources {
		if err := cmd.Kubectl().ApplyFile(resource).Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) Teardown(ctx context.Context) error {
	panic("implement me")
}


