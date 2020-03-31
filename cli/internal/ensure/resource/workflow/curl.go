package workflow

import (
	"context"
	"fmt"
	"io"
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
	DefaultMethod = "GET"
	DefaultPortForwardPort = 8080
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
	Method                string            `json:"method" valet:"default=GET"`
	RequestBody           string            `json:"body"`
	ResponseBody          string            `json:"responseBody"`
	ResponseBodySubstring string            `json:"responseBodySubstring"`
	Service               *ServiceRef       `json:"service"`
	PortForward           *PortForward      `json:"portForward"`
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
	delay, err := time.ParseDuration(c.Delay)
	if err != nil {
		return err
	}
	fullUrl, err := c.GetUrl(ctx, input)
	if err != nil {
		return err
	}

	var portForwardCmd *cmd.CommandStreamHandler
	if c.PortForward != nil {
		handler, err := c.PortForward.Initiate(ctx, input)
		if err != nil {
			return err
		}
		portForwardCmd = handler

		go func() {
			_ = handler.StreamHelper(nil)
		}()

		cmd.Stdout().Println("Initiated port forward")
	}

	cmd.Stdout().Println("Curling %s: {host: %s, headers: %v, expectedStatus: %d}", fullUrl, c.Host, c.Headers, c.StatusCode)

	curlErr := retry.Do(func() error {
		req, err := c.GetHttpRequest(fullUrl)
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

		if c.ResponseBodySubstring != "" && !strings.Contains(strings.TrimSpace(responseBody), strings.TrimSpace(c.ResponseBodySubstring)) {
			return UnexpectedResponseBodyError(responseBody)
		}

		cmd.Stdout().Println("Curl successful")
		return nil
	}, retry.Delay(delay), retry.Attempts(uint(c.Attempts)), retry.DelayType(retry.FixedDelay), retry.LastErrorOnly(true))

	if portForwardCmd != nil {
		_ = portForwardCmd.Process.Process.Kill()
	}
	return curlErr
}

func (c *Curl) GetUrl(ctx context.Context, input render.InputParams) (string, error) {
	if c.Service != nil {
		ipAndPort, err := c.Service.getAddress(ctx, input)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s://%s%s", c.Service.Port, ipAndPort, c.Path), nil
	} else if c.PortForward != nil {
		return fmt.Sprintf("http://localhost:%d%s", c.PortForward.Port, c.Path), nil
	}
	return "", errors.Errorf("Must specify either service or portForward")
}

func (c *Curl) GetHttpRequest(url string) (*http.Request, error) {
	var body io.Reader
	if c.RequestBody != "" {
		body = strings.NewReader(c.RequestBody)
	}
	req, err := http.NewRequest(c.Method, url, body)
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

type PortForward struct {
	Namespace      string `json:"namespace" valet:"key=Namespace"`
	DeploymentName string `json:"deploymentName"`
	Port           int    `json:"port" valet:"default=8080"`
}

func (p *PortForward) Initiate(ctx context.Context, input render.InputParams) (*cmd.CommandStreamHandler, error) {
	port := fmt.Sprintf("%d", p.Port)
	deployment := fmt.Sprintf("deploy/%s", p.DeploymentName)
	kubectl := cmd.New().Kubectl().With("port-forward").Namespace(p.Namespace).With(deployment, port).Cmd()
	return input.Runner().Stream(ctx, kubectl)
}