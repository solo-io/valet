package workflow

import (
	"context"

	"github.com/ghodss/yaml"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/cluster"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

type Config struct {
	Docs

	Cluster      *cluster.Cluster `json:"cluster"`
	CleanupSteps []Step           `json:"cleanupSteps"`
	Steps        []Step           `json:"steps"`
	Flags        render.Flags     `json:"flags"`
	Values       render.Values    `json:"values"`
}

func (c *Config) Ensure(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(c.Values)
	input = input.MergeFlags(c.Flags)
	if c.Cluster != nil {
		if err := c.Cluster.Ensure(ctx, input); err != nil {
			return err
		}
	}
	workflow := Workflow{
		Steps:        c.Steps,
		CleanupSteps: c.CleanupSteps,
	}
	return workflow.Ensure(ctx, input)
}

func (c *Config) Teardown(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(c.Values)
	input = input.MergeFlags(c.Flags)
	if c.Cluster != nil {
		return c.Cluster.Teardown(ctx, input)
	}
	workflow := Workflow{
		Steps:        c.Steps,
		CleanupSteps: c.CleanupSteps,
	}
	return workflow.Teardown(ctx, input)
}

func (c *Config) Document(ctx context.Context, input render.InputParams, section *Section) error {
	input = input.MergeValues(c.Values)
	input = input.MergeFlags(c.Flags)
	workflow := Workflow{
		Steps:        c.Steps,
		CleanupSteps: c.CleanupSteps,
	}
	workflow.Title = c.Title
	workflow.Description = c.Description

	return workflow.Document(ctx, input, section)
}

func LoadConfig(registry, path string, input render.InputParams) (*Config, error) {
	var c Config

	b, err := input.LoadFile(registry, path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict([]byte(b), &c, yaml.DisallowUnknownFields); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file '%s': %s", path, err.Error())
		return nil, err
	}

	return &c, nil
}
