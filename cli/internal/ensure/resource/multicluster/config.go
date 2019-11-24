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
		for i, cluster := range c.Clusters {
			cluster := cluster
			i := i
			eg.Go(func() error {
				return cluster.Ensure(cmd.NewPrinterContext(ctx, uint8(i), ""), input.DeepCopy())
			})
		}

		return eg.Wait()
	}

	for i, cluster := range c.Clusters {
		if err := cluster.Ensure(cmd.NewPrinterContext(ctx, uint8(i), ""), input.DeepCopy()); err != nil {
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
		for i, cluster := range c.Clusters {
			cluster := cluster
			i := i
			eg.Go(func() error {
				return cluster.Teardown(cmd.NewPrinterContext(ctx, uint8(i), ""), input.DeepCopy())
			})
		}

		return eg.Wait()
	}

	for i, cluster := range c.Clusters {
		if err := cluster.Teardown(cmd.NewPrinterContext(ctx, uint8(i), ""), input.DeepCopy()); err != nil {
			return err
		}
	}
	return nil
}

func LoadConfig(ctx context.Context, registry, path string, input render.InputParams) (*Config, error) {
	var c Config

	b, err := input.LoadFile(ctx, registry, path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict([]byte(b), &c); err != nil {
		cmd.Stderr(context.TODO()).Printf("Failed to unmarshal file '%s': %s", path, err.Error())
		return nil, err
	}

	return &c, nil
}
