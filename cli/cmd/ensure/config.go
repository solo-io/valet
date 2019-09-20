package ensure

import (
	"bytes"
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/options"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/url"
)

type ClusterConfig struct {
	Cluster   *options.Cluster   `yaml:"cluster"`
	Gloo      *options.Gloo      `yaml:"gloo"`
	Workflows []options.Workflow `yaml:"workflows"`
	Demos     *options.Demos     `yaml:"demos"`
	Resources []string           `yaml:"resources"`
}

func LoadConfig(ctx context.Context, path string) (*ClusterConfig, error) {
	var c ClusterConfig

	bytes, err := loadBytesFromPath(ctx, path)
	if err != nil {
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

func loadBytesFromPath(ctx context.Context, path string) ([]byte, error) {
	if isValidUrl(path) {
		bytes, err := loadBytesFromUrl(path)
		if err == nil {
			return bytes, nil
		}
		contextutils.LoggerFrom(ctx).Warnw("Could not read url, trying to read file",
			zap.Error(err),
			zap.String("path", path))
	}

	osClient := osutils.NewOsClient()
	bytes, err := osClient.ReadFile(path)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not read file",
			zap.Error(err),
			zap.String("path", path))
		return nil, err
	}
	return bytes, nil
}

func loadBytesFromUrl(path string) ([]byte, error) {
	// Get the data
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// isValidUrl tests a string to determine if it is a url or not.
func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}
