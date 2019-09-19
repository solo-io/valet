package workflow

import (
	"bytes"
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/options"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os/exec"
)

func EnsureWorkflow(ctx context.Context, workflow options.Workflow, glooUrl string) error {
	contextutils.LoggerFrom(ctx).Infow("Starting workflow", zap.String("name", workflow.Name))
	for _, step := range workflow.Steps {
		if err := runStep(ctx, step, glooUrl); err != nil {
			return err
		}
	}
	return nil
}

func runStep(ctx context.Context, step options.Step, glooUrl string) error {
	if step.Apply != "" {
		if err := apply(ctx, step.Apply); err != nil {
			return err
		}
	}

	if step.Curl != nil {
		if err := curl(ctx, step.Curl, glooUrl); err != nil {
			return err
		}
	}
	return nil
}

func apply(ctx context.Context, kubeYaml string) error {
	contextutils.LoggerFrom(ctx).Infow("Applying yaml", zap.String("kubeYaml", kubeYaml))
	cmd := exec.Command("kubectl", "apply", "-f", kubeYaml)
	out, err := cmd.CombinedOutput()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw(string(out), zap.Error(err), zap.Any("kubeYaml", kubeYaml))
		return err
	}
	return nil
}

func curl(ctx context.Context, curl *options.Curl, glooUrl string) error {
	fullUrl := fmt.Sprintf("%s%s", glooUrl, curl.Path)
	body := bytes.NewReader([]byte(glooUrl))
	req, err := http.NewRequest("GET", fullUrl, body)
	if err != nil {
		return err
	}

	req.Header = make(http.Header)
	for k, v := range curl.Headers {
		req.Header[k] = []string{v}
	}

	if curl.Host != "" {
		req.Host = curl.Host
	}

	contextutils.LoggerFrom(ctx).Infow("Curling",
		zap.String("url", fullUrl),
		zap.Any("headers", req.Header),
		zap.String("host", req.Host))

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

	if curl.StatusCode != resp.StatusCode {
		return errors.Errorf("Unexpected status code %s", resp.StatusCode)
	} else {
		contextutils.LoggerFrom(ctx).Infow("Curl got expected status code", zap.Int("statusCode", resp.StatusCode))
	}
	return nil
}
