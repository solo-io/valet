package gloo

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/testutils/kube"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

type GlooEnsurer interface {
	Install(ctx context.Context, config options.Gloo, localPathToGlooctl string) error
}

var _ GlooEnsurer = new(glooEnsurer)

func NewGlooEnsurer() *glooEnsurer {
	return &glooEnsurer{}
}

type glooEnsurer struct {
}

func (g *glooEnsurer) Install(ctx context.Context, config options.Gloo, localPathToGlooctl string) error {
	glooInstalled, err := checkForGlooInstall(ctx, config)
	if err != nil {
		return err
	}
	if glooInstalled {
		contextutils.LoggerFrom(ctx).Infow("Gloo is installed at the desired version")
		return nil
	}
	err = install(ctx, localPathToGlooctl, config)
	if err != nil {
		return err
	}
	return waitUntilPodsRunning(ctx, config)
}

func waitUntilPodsRunning(ctx context.Context, config options.Gloo) error {
	contextutils.LoggerFrom(ctx).Infow("Waiting for pods")
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return err
	}
	pods := kubeClient.CoreV1().Pods(config.Namespace)
	podsReady := func() (bool, error) {
		list, err := pods.List(v1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, pod := range list.Items {
			var podReady bool
			for _, cond := range pod.Status.Conditions {
				if cond.Type == v12.ContainersReady && cond.Status == v12.ConditionTrue {
					podReady = true
					break
				}
			}
			if !podReady {
				return false, nil
			}
		}
		return true, nil
	}
	failed := time.After(5 * time.Minute)
	notYetRunning := make(map[string]struct{})
	for {
		select {
		case <-failed:
			contextutils.LoggerFrom(ctx).Errorf("timed out waiting for pods to come online: %v", notYetRunning)
			return errors.Errorf("timed out waiting for pods to come online: %v", notYetRunning)
		case <-time.After(time.Second / 2):
			notYetRunning = make(map[string]struct{})
			ready, err := podsReady()
			if err != nil {
				contextutils.LoggerFrom(ctx).Errorw("error checking for ready pods", zap.Error(err))
				return err
			}
			if ready {
				contextutils.LoggerFrom(ctx).Infow("gloo is ready")
				return nil
			}
		}
	}
}

func checkForGlooInstall(ctx context.Context, config options.Gloo) (bool, error) {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return false, err
	}
	ns, err := kubeClient.CoreV1().Namespaces().Get(config.Namespace, v1.GetOptions{})
	if err != nil {
		if kubeerrs.IsNotFound(err) {
			return false, nil
		}
		contextutils.LoggerFrom(ctx).Errorw("Error trying to get namespace", zap.Error(err), zap.String("ns", config.Namespace))
		return false, err
	}
	if ns.Status.Phase != v12.NamespaceActive {
		contextutils.LoggerFrom(ctx).Errorw("Namespace is not active", zap.Any("phase", ns.Status.Phase))
	}
	pods, err := kubeClient.CoreV1().Pods(config.Namespace).List(v1.ListOptions{LabelSelector: "gloo"})
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error listing pods", zap.Error(err))
		return false, err
	}
	if len(pods.Items) == 0 {
		contextutils.LoggerFrom(ctx).Infow("No Gloo pods")
		return false, nil
	}
	for _, pod := range pods.Items {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == v12.ContainersReady && cond.Status != v12.ConditionTrue {
				contextutils.LoggerFrom(ctx).Infow("Gloo pods not ready")
				return false, nil
			}
		}
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if strings.Contains(container.Image, config.Version[1:]) {
				return true, nil
			}
		}
	}
	contextutils.LoggerFrom(ctx).Warnw("Detected Gloo install, but did not find any containers with the expected version",
		zap.String("expected", config.Version[1:]))
	return false, nil
}

func install(ctx context.Context, fullPath string, config options.Gloo) error {
	contextutils.LoggerFrom(ctx).Infow("Running glooctl install")
	args := []string{"install", "gateway"}
	if config.Enterprise {
		args = append(args, "--license-key", config.LicenseKey)
	}
	out, err := internal.ExecuteCmd(fullPath, args...)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to install gloo",
			zap.Error(err),
			zap.String("out", out))
	}
	return err
}
