package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var _ Resource = new(Namespace)

const (
	ns = "ns"
)

type Namespace struct {
	Name   string            `yaml:"name"`
	Labels map[string]string `yaml:"labels"`
}

func (n *Namespace) Ensure(ctx context.Context, command cmd.Factory) error {
	err := command.Kubectl().Create(ns).WithName(n.Name).DryRunAndApply(ctx, command)
	if err != nil {
		return err
	}
	for k, v := range n.Labels {
		labelString := fmt.Sprintf("%s=%s", k, v)
		if err := command.Kubectl().With("label", "ns", n.Name, labelString, "--overwrite").Cmd().Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (n *Namespace) Teardown(ctx context.Context, command cmd.Factory) error {
	return command.Kubectl().Delete(ns).WithName(n.Name).Cmd().Run(ctx)
}
