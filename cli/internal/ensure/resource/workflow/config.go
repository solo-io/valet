package workflow

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/cluster"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Cluster      *cluster.Cluster `yaml:"cluster"`
	CleanupSteps []Step           `yaml:"cleanupSteps"`
	Steps        []Step           `yaml:"steps"`
	Flags        render.Flags     `yaml:"flags"`
	Values       render.Values    `yaml:"values"`
}

func (c *Config) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(c.Values)
	input = input.MergeFlags(c.Flags)
	if c.Cluster != nil {
		if err := c.Cluster.Ensure(ctx, input, command); err != nil {
			return err
		}
	}
	workflow := Workflow{
		Steps:        c.Steps,
		CleanupSteps: c.CleanupSteps,
	}
	return workflow.Ensure(ctx, input, command)
}

func (c *Config) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(c.Values)
	input = input.MergeFlags(c.Flags)
	if c.Cluster != nil {
		return c.Cluster.Teardown(ctx, input, command)
	}
	workflow := Workflow{
		Steps:        c.Steps,
		CleanupSteps: c.CleanupSteps,
	}
	return workflow.Teardown(ctx, input, command)
}

func LoadConfig(registry, path string, input render.InputParams) (*Config, error) {
	var c Config

	b, err := input.LoadFile(registry, path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict([]byte(b), &c); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file '%s': %s", path, err.Error())
		return nil, err
	}

	return &c, nil
}