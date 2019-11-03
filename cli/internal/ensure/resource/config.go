package resource

import (
	"bytes"
	"context"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/url"
)

type Config struct {
	Cluster      *Cluster          `yaml:"cluster"`
	Applications []ApplicationRef  `yaml:"applications"`
	Workflows    []Workflow        `yaml:"workflows"`
	Flags        []string          `yaml:"flags"`
	Values       map[string]string `yaml:"values"`
}

func (c *Config) Ensure(ctx context.Context, command cmd.Factory) error {
	if c.Cluster != nil {
		if err := c.Cluster.Ensure(ctx, command); err != nil {
			return err
		}
	}

	for _, application := range c.Applications {
		if c.Values != nil {
			application.updateWithValues(c.Values)
		}
		if c.Flags != nil {
			application.Flags = append(application.Flags, c.Flags...)
		}
		if err := application.Ensure(ctx, command); err != nil {
			return err
		}
	}
	for _, workflow := range c.Workflows {
		//workflow.URL = proxyUrl
		if err := workflow.Ensure(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) Teardown(ctx context.Context, command cmd.Factory) error {
	if c.Cluster != nil {
		return c.Cluster.Teardown(ctx, command)
	}
	for _, application := range c.Applications {
		if err := application.Teardown(ctx, command); err != nil {
			return err
		}
	}
	return nil
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
