package resource

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Cluster      *Cluster `yaml:"cluster"`
	CleanupSteps []Step   `yaml:"cleanupSteps"`
	Steps        []Step   `yaml:"steps"`
	Flags        Flags    `yaml:"flags"`
	Values       Values   `yaml:"values"`
}

func (c *Config) Ensure(ctx context.Context, input InputParams, command cmd.Factory) error {
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

func (c *Config) Teardown(ctx context.Context, input InputParams, command cmd.Factory) error {
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

func LoadConfig(path string) (*Config, error) {
	var c Config

	b, err := loadBytesFromPath(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, &c); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file '%s': %s", path, err.Error())
		return nil, err
	}

	return &c, nil
}

func loadBytesFromPath(path string) ([]byte, error) {
	if isValidUrl(path) {
		contents, err := loadBytesFromUrl(path)
		if err == nil {
			return contents, nil
		}
	}

	osClient := osutils.NewOsClient()
	contents, err := osClient.ReadFile(path)
	if err != nil {
		cmd.Stderr().Println("Failed to read file '%s': %s", path, err.Error())
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
