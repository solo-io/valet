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

func (c *Config) Ensure(ctx context.Context, command cmd.Factory) error {
	if c.Cluster != nil {
		if err := c.Cluster.Ensure(ctx, command); err != nil {
			return err
		}
	}
	for _, ns := range c.Namespaces {
		if err := ns.Ensure(ctx, command); err != nil {
			return err
		}
	}
	for _, secret := range c.Secrets {
		if err := secret.Ensure(ctx, command); err != nil {
			return err
		}
	}
	if c.CertManager != nil {
		if err := c.CertManager.Ensure(ctx, command); err != nil {
			return err
		}
	}
	proxyUrl := ""
	if c.Gloo != nil {
		if err := c.Gloo.Ensure(ctx, command); err != nil {
			return err
		}
		proxyAddress, err := c.Gloo.GetProxyAddress(ctx, command)
		if err != nil {
			return err
		}
		proxyUrl = proxyAddress
	}
	if c.ServiceMeshHub != nil {
		if err := c.ServiceMeshHub.Ensure(ctx, command); err != nil {
			return err
		}
	}

	for _, workflow := range c.Workflows {
		workflow.URL = proxyUrl
		if err := workflow.Ensure(ctx, command); err != nil {
			return err
		}
	}
	if c.Demos != nil {
		if err := c.Demos.Ensure(ctx, command); err != nil {
			return err
		}
	}
	for _, resource := range c.Resources {
		if err := command.Kubectl().ApplyFile(resource).Cmd().Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) Teardown(ctx context.Context, command cmd.Factory) error {
	if c.Cluster != nil {
		return c.Cluster.Teardown(ctx, command)
	}
	if c.CertManager != nil {
		if err := c.CertManager.Teardown(ctx, command); err != nil {
			return err
		}
	}
	if c.Demos != nil {
		if err := c.Demos.Teardown(ctx, command); err != nil {
			return err
		}
	}
	for _, namespace := range c.Namespaces {
		if err := namespace.Teardown(ctx, command); err != nil {
			return err
		}
	}
	if c.ServiceMeshHub != nil {
		if err := c.ServiceMeshHub.Teardown(ctx, command); err != nil {
			return err
		}
	}
	for _, secret := range c.Secrets {
		if err := secret.Teardown(ctx, command); err != nil {
			return err
		}
	}
	if c.Gloo != nil {
		if err := c.Gloo.Teardown(ctx, command); err != nil {
			return err
		}
	}
	return nil
}


