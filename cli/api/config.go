package api

import (
	"bytes"
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/url"
)

func LoadConfig(ctx context.Context, path string) (*resource.Config, error) {
	var c resource.Config

	b, err := loadBytesFromPath(ctx, path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, &c); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to unmarshal file",
			zap.Error(err),
			zap.String("path", path),
			zap.ByteString("bytes", b))
		return nil, err
	}

	return &c, nil
}

func LoadWorkflow(ctx context.Context, path string) (*resource.Workflow, error) {
	var w resource.Workflow

	b, err := loadBytesFromPath(ctx, path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, &w); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to unmarshal file",
			zap.Error(err),
			zap.String("path", path),
			zap.ByteString("bytes", b))
		return nil, err
	}

	return &w, nil
}

func loadBytesFromPath(ctx context.Context, path string) ([]byte, error) {
	if isValidUrl(path) {
		contents, err := loadBytesFromUrl(path)
		if err == nil {
			return contents, nil
		}
	}

	osClient := osutils.NewOsClient()
	contents, err := osClient.ReadFile(path)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Could not read file",
			zap.Error(err),
			zap.String("path", path))
		return nil, err
	}
	return contents, nil
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
