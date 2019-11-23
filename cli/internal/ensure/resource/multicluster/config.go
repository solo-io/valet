package multicluster

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Clusters      []*Ref        `yaml:"clusters"`
	Values        render.Values `yaml:"values"`
	Flags         render.Flags  `yaml:"flags"`
	RunInParallel bool          `yaml:"runInParallel"`
}


func (c *Config) Ensure(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(c.Values)
	input = input.MergeFlags(c.Flags)

	eg := errgroup.Group{}
	if c.RunInParallel {
		for _, cluster := range c.Clusters {
			cluster := cluster
			eg.Go(func() error {
				return cluster.Ensure(ctx, input)
			})
		}

		return eg.Wait()
	}

	for _, cluster := range c.Clusters {
		if err := cluster.Ensure(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) Teardown(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(c.Values)
	input = input.MergeFlags(c.Flags)

	eg := errgroup.Group{}
	if c.RunInParallel {
		for _, cluster := range c.Clusters {
			cluster := cluster
			eg.Go(func() error {
				return cluster.Teardown(ctx, input)
			})
		}

		return eg.Wait()
	}

	for _, cluster := range c.Clusters {
		if err := cluster.Teardown(ctx, input); err != nil {
			return err
		}
	}
	return nil
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
