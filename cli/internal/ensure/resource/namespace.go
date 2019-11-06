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
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

func (n *Namespace) Ensure(ctx context.Context, inputs InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Ensuring namespace %s", n.Name)
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
	for k, v := range n.Annotations {
		annotationString := fmt.Sprintf("%s=%s", k, v)
		if err := command.Kubectl().With("annotate", "ns", n.Name, annotationString, "--overwrite").Cmd().Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (n *Namespace) Teardown(ctx context.Context, inputs InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Tearing down namespace %s", n.Name)
	return command.Kubectl().Delete(ns).WithName(n.Name).Cmd().Run(ctx)
}
