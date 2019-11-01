package resource

import (
	"bytes"
	"context"
	"fmt"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"io"
	"net/http"
)

var (
	NoUrlSetError = errors.Errorf("no url set")
	UnexpectedStatusCodeError = func(statusCode int) error {
		return errors.Errorf("Curl got unexpected status code %d", statusCode)
	}
)

type Curl struct {
	Path       string            `yaml:"path"`
	Host       string            `yaml:"host"`
	Headers    map[string]string `yaml:"headers"`
	StatusCode int               `yaml:"statusCode"`
	URL        string
}

func (c *Curl) Ensure(ctx context.Context, command cmd.Factory) error {
	if c.URL == "" {
		return NoUrlSetError
	}
	return c.doCurl(ctx, command)
}

func (c *Curl) Teardown(ctx context.Context, command cmd.Factory) error {
	return nil
}

func (c *Curl) doCurl(ctx context.Context, command cmd.Factory) error {
	fullUrl := fmt.Sprintf("%s%s", c.URL, c.Path)
	body := bytes.NewReader([]byte(c.URL))
	req, err := http.NewRequest("GET", fullUrl, body)
	if err != nil {
		return err
	}

	req.Header = make(http.Header)
	for k, v := range c.Headers {
		req.Header[k] = []string{v}
	}

	if c.Host != "" {
		req.Host = c.Host
	}
	cmd.Stdout().Println("Curling %s: {host: %s, headers: %v, expectedStatus: %d}", fullUrl, req.Host, req.Header, c.StatusCode)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	p := new(bytes.Buffer)
	_, err = io.Copy(p, resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if c.StatusCode != resp.StatusCode {
		cmd.Stderr().Println("Curl got unexpected status code %d", resp.StatusCode)
		return UnexpectedStatusCodeError(resp.StatusCode)
	} else {
		cmd.Stdout().Println("Curl got expected status code %d", resp.StatusCode)
	}
	return nil
}
