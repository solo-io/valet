package workflow

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	errors "github.com/rotisserie/eris"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	DefaultCurlDelay    = "1s"
	DefaultCurlAttempts = 10
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
	Path                  string            `json:"path"`
	Host                  string            `json:"host"`
	Headers               map[string]string `json:"headers"`
	StatusCode            int               `json:"statusCode"`
	ResponseBody          string            `json:"responseBody"`
	ResponseBodySubstring string            `json:"responseBodySubstring"`
	Service               ServiceRef        `json:"service"`
	Attempts              int               `json:"attempts" valet:"default=10"`
	Delay                 string            `json:"delay" valet:"default=1s"`
}

func (c *Curl) Run(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(c); err != nil {
		return err
	}
	return c.doCurl(ctx, input)
}

func (c *Curl) Ensure(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(c); err != nil {
		return err
	}
	return c.doCurl(ctx, input)
}

func (c *Curl) Teardown(ctx context.Context, input render.InputParams) error {
	return nil
}

func (c *Curl) doCurl(ctx context.Context, input render.InputParams) error {
	ip, err := c.Service.getAddress(ctx, input)
	if err != nil {
		return err
	}
	delay, err := time.ParseDuration(c.Delay)
	fullUrl := c.GetUrl(ip)
	cmd.Stdout().Println("Curling %s: {host: %s, headers: %v, expectedStatus: %d}", fullUrl, c.Host, c.Headers, c.StatusCode)

	return retry.Do(func() error {
		req, err := c.GetHttpRequest(ip)
		if err != nil {
			return err
		}
		responseBody, statusCode, err := input.Runner().Request(ctx, req)
		if err != nil {
			return err
		}
		if c.StatusCode != statusCode {
			return UnexpectedStatusCodeError(statusCode)
		}
		if c.ResponseBody != "" && strings.TrimSpace(responseBody) != strings.TrimSpace(c.ResponseBody) {
			return UnexpectedResponseBodyError(responseBody)
		}

		if c.ResponseBodySubstring != "" && !strings.Contains(strings.TrimSpace(responseBody),strings.TrimSpace(c.ResponseBodySubstring)) {
			return UnexpectedResponseBodyError(responseBody)
		}

		cmd.Stdout().Println("Curl successful")
		return nil
	}, retry.Delay(delay), retry.Attempts(uint(c.Attempts)), retry.DelayType(retry.FixedDelay), retry.LastErrorOnly(true))

}

func (c *Curl) GetUrl(ip string) string {
	return fmt.Sprintf("%s://%s%s", c.Service.Port, ip, c.Path)
}

func (c *Curl) GetHttpRequest(ip string) (*http.Request, error) {
	url := c.GetUrl(ip)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = make(http.Header)
	for k, v := range c.Headers {
		req.Header[k] = []string{v}
	}

	if c.Host != "" {
		req.Host = c.Host
	}
	return req, nil
}
