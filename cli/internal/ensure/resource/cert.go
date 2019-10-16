package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Cert struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Domain    string `yaml:"domain"`
}

func (c *Cert) Ensure(ctx context.Context, command cmd.Factory) error {
	return command.Kubectl().ApplyStdIn(c.GetCertYaml()).Cmd().Run(ctx)
}

func (c *Cert) Teardown(ctx context.Context, command cmd.Factory) error {
	return command.Kubectl().DeleteStdIn(c.GetCertYaml()).Cmd().Run(ctx)
}

func (c *Cert) GetCertYaml() string {
	return internal.CreateCertString(c.Name, c.Namespace, c.Domain)
}
