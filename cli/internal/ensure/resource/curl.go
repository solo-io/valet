package resource

import (
	"bytes"
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
)

var (
	NoUrlSetError = errors.Errorf("no url set")
)

type Curl struct {
	Path       string            `yaml:"path"`
	Host       string            `yaml:"host"`
	Headers    map[string]string `yaml:"headers"`
	StatusCode int               `yaml:"statusCode"`
	URL        string
}

func (c *Curl) Ensure(ctx context.Context) error {
	if c.URL == "" {
		return NoUrlSetError
	}
	return c.doCurl(ctx)
}

func (c *Curl) Teardown(ctx context.Context) error {
	return nil
}

func (c *Curl) doCurl(ctx context.Context) error {
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

	contextutils.LoggerFrom(ctx).Infow("Curling",
		zap.String("url", fullUrl),
		zap.Any("headers", req.Header),
		zap.String("host", req.Host),
		zap.Int("expectedStatus", c.StatusCode))

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
		contextutils.LoggerFrom(ctx).Warnw("Curl got unexpected status code", zap.String("url", fullUrl), zap.String("host", req.Host), zap.Int("expected", curl.StatusCode), zap.Int("actual", resp.StatusCode))
		return errors.Errorf("Unexpected status code %d", resp.StatusCode)
	} else {
		contextutils.LoggerFrom(ctx).Infow("Curl got expected status code", zap.Int("statusCode", resp.StatusCode))
	}
	return nil
}
