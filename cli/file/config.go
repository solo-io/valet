package file

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/kube-cluster/cli/options"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Cluster *options.Cluster `yaml:"cluster"`
	Gloo    *options.Gloo    `yaml:"gloo"`
	Demos   *options.Demos   `yaml:"demos"`
}

func LoadConfig(ctx context.Context, path string) (*Config, error) {
	var c Config

	osClient := osutils.NewOsClient()
	bytes, err := osClient.ReadFile(path)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not read file",
			zap.Error(err),
			zap.String("path", path))
		return nil, err
	}

	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to unmarshal file",
			zap.Error(err),
			zap.String("path", path),
			zap.ByteString("bytes", bytes))
		return nil, err
	}

	return &c, nil
}