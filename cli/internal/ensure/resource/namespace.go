package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ Resource = new(Namespace)

const (
	ns = "ns"
)

type Namespace struct {
	Name string `yaml:"name"`
}

func (n *Namespace) Ensure(ctx context.Context) error {
	return cmd.Kubectl().Create(ns).WithName(n.Name).DryRunAndApply(ctx)
}

func (n *Namespace) Teardown(ctx context.Context) error {
	return cmd.Kubectl().Delete(ns).WithName(n.Name).Run(ctx)
}
