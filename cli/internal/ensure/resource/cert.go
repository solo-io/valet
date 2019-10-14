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

func (c *Cert) Ensure(ctx context.Context) error {
	cert := internal.CreateCert(c.Name, c.Namespace, c.Domain)
	return cmd.Kubectl().ApplyStdIn(cert).Run()
}

func (c *Cert) Teardown(ctx context.Context) error {
	cert := internal.CreateCert(c.Name, c.Namespace, c.Domain)
	return cmd.Kubectl().DeleteStdIn(cert).Run()
}
