package multicluster

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/cluster"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"github.com/solo-io/valet/cli/internal/ensure/resource/workflow"
	"gopkg.in/yaml.v2"
)

type Workflow struct {
	Cluster      *cluster.Cluster `yaml:"cluster"`
	CleanupSteps []workflow.Step  `yaml:"cleanupSteps"`
	Steps        []workflow.Step  `yaml:"steps"`
	Flags        render.Flags     `yaml:"flags"`
	Values       render.Values    `yaml:"values"`
}

func (c *Workflow) Ensure(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(c.Values)
	input = input.MergeFlags(c.Flags)

	if c.Cluster != nil {
		if err := input.RenderFields(c.Cluster); err != nil {
			return err
		}
		if err := c.Cluster.Ensure(ctx, input); err != nil {
			return err
		}
	}
	workflow := workflow.Workflow{
		Steps:        c.Steps,
		CleanupSteps: c.CleanupSteps,
	}
	return workflow.Ensure(ctx, input)
}

func (c *Workflow) Teardown(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(c.Values)
	input = input.MergeFlags(c.Flags)
	if c.Cluster != nil {
		if err := input.RenderFields(c.Cluster); err != nil {
			return err
		}
		return c.Cluster.Teardown(ctx, input)
	}
	workflow := workflow.Workflow{
		Steps:        c.Steps,
		CleanupSteps: c.CleanupSteps,
	}
	return workflow.Teardown(ctx, input)
}

func LoadClusterWorkflow(ctx context.Context, registry, path string, input render.InputParams) (*Workflow, error) {
	var c Workflow

	b, err := input.LoadFile(ctx, registry, path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict([]byte(b), &c); err != nil {
		cmd.Stderr(ctx).Println("Failed to unmarshal file '%s': %s", path, err.Error())
		return nil, err
	}

	return &c, nil
}
