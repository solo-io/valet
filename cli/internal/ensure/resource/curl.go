package resource

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var (
	UnexpectedStatusCodeError = func(statusCode int) error {
		return errors.Errorf("Curl got unexpected status code %d", statusCode)
	}
	UnexpectedResponseBodyError = func(responseBody string) error {
		return errors.Errorf("Curl got unexpected response body:\n%s", responseBody)
	}
)

type Curl struct {
	Path         string            `yaml:"path"`
	Host         string            `yaml:"host"`
	Protocol     string            `yaml:"protocol" valet:"default=http"`
	Headers      map[string]string `yaml:"headers"`
	StatusCode   int               `yaml:"statusCode"`
	ResponseBody string            `yaml:"responseBody"`
	Service      ServiceRef        `yaml:"service"`
}

func (c *Curl) Ensure(ctx context.Context, input InputParams, command cmd.Factory) error {
	if err := input.Values.RenderFields(c); err != nil {
		return err
	}
	return c.doCurl(ctx, input, command)
}

func (c *Curl) Teardown(ctx context.Context, input InputParams, command cmd.Factory) error {
	return nil
}

func (c *Curl) doCurl(ctx context.Context, input InputParams, command cmd.Factory) error {
	ip, err := c.Service.getAddress(ctx, input, command)
	if err != nil {
		return err
	}
	fullUrl := fmt.Sprintf("%s://%s%s", c.Protocol, ip, c.Path)
	cmd.Stdout().Println("Curling %s: {host: %s, headers: %v, expectedStatus: %d}", fullUrl, c.Host, c.Headers, c.StatusCode)

	return retry.Do(func() error {
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
		httpClient := &http.Client{
			Timeout: time.Second * 1,
		}
		resp, err := httpClient.Do(req)
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
			return UnexpectedStatusCodeError(resp.StatusCode)
		}
		if c.ResponseBody != "" && strings.TrimSpace(p.String()) != strings.TrimSpace(c.ResponseBody) {
			return UnexpectedResponseBodyError(p.String())
		}

		cmd.Stdout().Println("Curl successful")
		return nil
	}, retry.Delay(time.Second), retry.Attempts(10), retry.DelayType(retry.FixedDelay), retry.LastErrorOnly(true))

}
