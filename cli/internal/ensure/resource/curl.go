package resource

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var (
	UnexpectedStatusCodeError = func(statusCode int) error {
		return errors.Errorf("Curl got unexpected status code %d", statusCode)
	}
)

type Curl struct {
	Path       string            `yaml:"path"`
	Host       string            `yaml:"host"`
	Headers    map[string]string `yaml:"headers"`
	StatusCode int               `yaml:"statusCode"`
	Service    ServiceRef        `yaml:"service"`
}

func (c *Curl) Ensure(ctx context.Context, command cmd.Factory) error {
	return c.doCurl(ctx, command)
}

func (c *Curl) Teardown(ctx context.Context, command cmd.Factory) error {
	return nil
}

func (c *Curl) doCurl(ctx context.Context, command cmd.Factory) error {
	ip, err := c.Service.getIpAddress(ctx, command)
	if err != nil {
		return err
	}
	fullUrl := fmt.Sprintf("%s%s", ip, c.Path)
	body := bytes.NewReader([]byte(ip))
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
