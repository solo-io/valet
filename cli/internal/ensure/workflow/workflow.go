package workflow

import (
	"bytes"
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/internal/ensure/gloo"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

var _ WorkflowRunner = new(workflowRunner)

type WorkflowRunner interface {
	Run(ctx context.Context, workflowPath string) error
}

func NewWorkflowRunner(gloo gloo.GlooManager) WorkflowRunner {
	return &workflowRunner{
		gloo: gloo,
		resources: make(map[string]bool),
	}
}

type workflowRunner struct {
	gloo    gloo.GlooManager

	glooUrl string
	resources map[string]bool
}

func (w *workflowRunner) Run(ctx context.Context, workflowPath string) error {
	contextutils.LoggerFrom(ctx).Infow("Loading workflow", zap.String("path", workflowPath))
	workflow, err := api.LoadWorkflow(ctx, workflowPath)
	if err != nil {
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("Starting workflow", zap.String("name", workflow.Name))
	for _, step := range workflow.Steps {
		if err := w.runStep(ctx, step); err != nil {
			return err
		}
	}
	contextutils.LoggerFrom(ctx).Infow("Workflow passed, cleaning up resources")
	if err := w.cleanupResources(ctx); err != nil {
		return err
	}
	return nil
}

func (w *workflowRunner) cleanupResources(ctx context.Context) error {
	for k, v  := range w.resources {
		if v {
			// resource was not deleted, clean up
			if err := kubectlDelete(ctx, k); err != nil {
				return err
			}
		}
		delete(w.resources, k)
	}
	return nil
}

func (w *workflowRunner) runStep(ctx context.Context, step api.Step) error {
	if step.Apply != "" {
		if err := apply(ctx, step.Apply); err != nil {
			return err
		}
		w.resources[step.Apply] = true
	}

	if step.Delete != "" {
		if err := kubectlDelete(ctx, step.Delete); err != nil {
			return err
		}
		w.resources[step.Delete] = false
	}

	if step.Curl != nil {
		if err := w.getGlooUrlAndCurlWithRetry(ctx, step.Curl, 5); err != nil {
			return err
		}
	}
	return nil
}

func kubectlDelete(ctx context.Context, kubeYaml string) error {
	contextutils.LoggerFrom(ctx).Infow("Deleting yaml", zap.String("kubeYaml", kubeYaml))
	cmd := exec.Command("kubectl", "delete", "-f", kubeYaml)
	out, err := cmd.CombinedOutput()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw(string(out), zap.Error(err), zap.Any("kubeYaml", kubeYaml))
		return err
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

func (w *workflowRunner) getGlooUrlAndCurlWithRetry(ctx context.Context, curl *api.Curl, retries int) error {
	if w.glooUrl == "" {
		glooUrl, err := w.getGlooUrl()
		if err != nil {
			return err
		}
		w.glooUrl = glooUrl
	}
	err := w.doCurl(ctx, curl)
	if err != nil {
		if retries > 0 {
			w.glooUrl = ""
			return w.getGlooUrlAndCurlWithRetry(ctx, curl, retries-1)
		}
	}
	return err
}

func (w *workflowRunner) getGlooUrl() (string, error) {
	out, err := w.gloo.Glooctl().Execute("proxy", "url")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (w *workflowRunner) doCurl(ctx context.Context, curl *api.Curl) error {
	fullUrl := fmt.Sprintf("%s%s", w.glooUrl, curl.Path)
	body := bytes.NewReader([]byte(w.glooUrl))
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
		return errors.Errorf("Unexpected status code %d", resp.StatusCode)
	} else {
		contextutils.LoggerFrom(ctx).Infow("Curl got expected status code", zap.Int("statusCode", resp.StatusCode))
	}
	return nil
}
