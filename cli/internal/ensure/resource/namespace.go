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

func (n *Namespace) Ensure(ctx context.Context, command cmd.Factory) error {
	return command.Kubectl().Create(ns).WithName(n.Name).DryRunAndApply(ctx, command)
}

func (n *Namespace) Teardown(ctx context.Context, command cmd.Factory) error {
	return command.Kubectl().Delete(ns).WithName(n.Name).Cmd().Run(ctx)
}
