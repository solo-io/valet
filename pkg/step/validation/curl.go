package validation

import (
	"fmt"
	"github.com/avast/retry-go"
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
	"io"
	"net/http"
	"strings"
	"time"
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
	StatusCode            int               `json:"statusCode" valet:"default=200"`
	Method                string            `json:"method" valet:"default=GET"`
	RequestBody           string            `json:"body"`
	ResponseBody          string            `json:"responseBody"`
	ResponseBodySubstring string            `json:"responseBodySubstring"`
	Service               *ServiceRef       `json:"service"`
	PortForward           *PortForward      `json:"portForward"`
	Attempts              int               `json:"attempts" valet:"default=10"`
	Delay                 string            `json:"delay" valet:"default=1s"`
}

func (c *Curl) Run(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(c, ctx.Runner); err != nil {
		return err
	}
	return c.doCurl(ctx, values)
}

func (c *Curl) GetDescription(ctx *api.WorkflowContext, values render.Values) (string, error) {
	if err := values.RenderFields(c, ctx.Runner); err != nil {
		return "", err
	}
	url, err := c.GetUrl(ctx, values)
	if err != nil {
		return "", err
	}
	str := fmt.Sprintf("Issuing http request\n%s %s", c.Method, url)
	if c.RequestBody != "" {
		str += fmt.Sprintf("\nBody: %s", c.RequestBody)
	}
	str += fmt.Sprintf("\nExpected status: %d", c.StatusCode)
	if c.ResponseBody != "" {
		str += fmt.Sprintf("\nExpected response: %s", c.ResponseBody)
	} else if c.ResponseBodySubstring != "" {
		str += fmt.Sprintf("\nExpected response substring: %s", c.ResponseBodySubstring)
	}
	return str, nil
}

func (c *Curl) GetDocs(ctx *api.WorkflowContext, options api.DocsOptions) (string, error) {
	panic("implement me")
}

func (c *Curl) doCurl(ctx *api.WorkflowContext, values render.Values) error {
	delay, err := time.ParseDuration(c.Delay)
	if err != nil {
		return err
	}
	fullUrl, err := c.GetUrl(ctx, values)
	if err != nil {
		return err
	}

	var portForwardCmd *cmd.CommandStreamHandler
	if c.PortForward != nil {
		handler, err := c.PortForward.Initiate(ctx, values)
		if err != nil {
			return err
		}
		portForwardCmd = handler

		go func() {
			_ = handler.StreamHelper(nil)
		}()

		cmd.Stdout().Println("Initiated port forward")
	}

	curlErr := retry.Do(func() error {
		req, err := c.GetHttpRequest(fullUrl)
		if err != nil {
			return err
		}
		responseBody, statusCode, err := ctx.Runner.Request(req)
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
		_ = ctx.Runner.Kill(portForwardCmd.Process.Process)
	}
	return curlErr
}

func (c *Curl) GetUrl(ctx *api.WorkflowContext, values render.Values) (string, error) {
	if c.Service != nil {
		ipAndPort, err := c.Service.getAddress(ctx, values)
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

func (p *PortForward) Initiate(ctx *api.WorkflowContext, values render.Values) (*cmd.CommandStreamHandler, error) {
	port := fmt.Sprintf("%d", p.Port)
	deployment := fmt.Sprintf("deploy/%s", p.DeploymentName)
	kubectl := cmd.New().Kubectl().With("port-forward").Namespace(p.Namespace).With(deployment, port).Cmd()
	return ctx.Runner.Stream(kubectl)
}

type ServiceRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace" valet:"key=Namespace"`
	Port      string `json:"port" valet:"default=http"`
}

func (s *ServiceRef) getAddress(ctx *api.WorkflowContext, values render.Values) (string, error) {
	if err := values.RenderFields(s, ctx.Runner); err != nil {
		return "", err
	}
	return ctx.KubeClient.GetIngressHost(s.Name, s.Namespace, s.Port)
}

func (s *ServiceRef) getIp(ctx *api.WorkflowContext, values render.Values) (string, error) {
	url, err := s.getAddress(ctx, values)
	if err != nil {
		return "", err
	}
	parts := strings.Split(url, ":")
	if len(parts) <= 2 {
		return parts[0], nil
	}
	return "", errors.Errorf("Unexpected url %s", url)
}