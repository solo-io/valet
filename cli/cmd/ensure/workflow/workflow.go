package workflow

import (
	"bytes"
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

func EnsureWorkflow(top *options.Top, workflow options.Workflow) error {
	contextutils.LoggerFrom(top.Ctx).Infow("Starting workflow", zap.String("name", workflow.Name))
	for _, step := range workflow.Steps {
		if err := runStep(top, step); err != nil {
			return err
		}
	}
	return nil
}

func runStep(top *options.Top, step options.Step) error {
	if step.Apply != "" {
		if err := apply(top.Ctx, step.Apply); err != nil {
			return err
		}
	}

	if step.Curl != nil {
		if err := getGlooUrlAndCurlWithRetry(top, step.Curl, 5); err != nil {
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

func getGlooUrlAndCurlWithRetry(top *options.Top, curl *options.Curl, retries int) error {
	if top.GlooUrl == "" {
		glooUrl, err := getGlooUrl(top.LocalPathToGlooctl)
		if err != nil {
			return err
		}
		top.GlooUrl = glooUrl
	}

	err := doCurl(top, curl)
	if err != nil {
		if retries > 0 {
			top.GlooUrl = ""
			return getGlooUrlAndCurlWithRetry(top, curl, retries - 1)
		}
	}
	return err
}

func getGlooUrl(localPathToGlooctl string) (string, error) {
	out, err := internal.ExecuteCmd(localPathToGlooctl, "proxy", "url")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func doCurl(top *options.Top, curl *options.Curl) error {
	fullUrl := fmt.Sprintf("%s%s", top.GlooUrl, curl.Path)
	body := bytes.NewReader([]byte(top.GlooUrl))
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

	contextutils.LoggerFrom(top.Ctx).Infow("Curling",
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
		contextutils.LoggerFrom(top.Ctx).Infow("Curl got expected status code", zap.Int("statusCode", resp.StatusCode))
	}
	return nil
}
