package gloo

import (
	"context"
	"fmt"
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

const DefaultNamespace = "gloo-system"

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
	glooInstalled, err := checkForGlooInstall(ctx, config, localPathToGlooctl)
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
	pods := kubeClient.CoreV1().Pods(DefaultNamespace)
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

func checkForGlooInstall(ctx context.Context, config options.Gloo, localPathToGlooctl string) (bool, error) {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return false, err
	}
	ns, err := kubeClient.CoreV1().Namespaces().Get(DefaultNamespace, v1.GetOptions{})
	if err != nil {
		if kubeerrs.IsNotFound(err) {
			return false, nil
		}
		contextutils.LoggerFrom(ctx).Errorw("Error trying to get namespace", zap.Error(err), zap.String("ns", DefaultNamespace))
		return false, err
	}
	if ns.Status.Phase != v12.NamespaceActive {
		contextutils.LoggerFrom(ctx).Errorw("Namespace is not active", zap.Any("phase", ns.Status.Phase))
	}
	pods, err := kubeClient.CoreV1().Pods(DefaultNamespace).List(v1.ListOptions{LabelSelector: "gloo"})
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
	version := config.Version
	if !config.ValetArtifacts && config.LocalArtifactDir != "" {
		version = version[1:]
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if strings.Contains(container.Image, version) {
				return true, nil
			}
		}
	}
	contextutils.LoggerFrom(ctx).Warnw("Detected Gloo install, but did not find any containers with the expected version",
		zap.String("expected", version))
	return false, uninstallAll(ctx, localPathToGlooctl)
}

func install(ctx context.Context, fullPath string, config options.Gloo) error {
	contextutils.LoggerFrom(ctx).Infow("Running glooctl install")
	args := []string{"install", "gateway"}
	if config.Enterprise {
		args = append(args, "--license-key", config.LicenseKey)
	}

	if config.LocalArtifactDir != "" {
		helmChart := fmt.Sprintf("_artifacts/gloo-%s.tgz", config.Version)
		args = append(args, "-f", helmChart)
		contextutils.LoggerFrom(ctx).Infow("Using helm chart from local artifacts", zap.String("helmChart", helmChart))
	} else if config.ValetArtifacts {
		helmChart := fmt.Sprintf("https://storage.googleapis.com/valet/artifacts/gloo/%s/gloo-%s.tgz", config.Version, config.Version)
		args = append(args, "-f", helmChart)
		contextutils.LoggerFrom(ctx).Infow("Using helm chart from valet artifacts", zap.String("helmChart", helmChart))
	}
	out, err := internal.ExecuteCmd(fullPath, args...)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to install gloo",
			zap.Error(err),
			zap.String("out", out))
	}
	return err
}

func uninstallAll(ctx context.Context, fullPath string) error {
	contextutils.LoggerFrom(ctx).Infow("Uninstalling existing gloo")
	args := []string{"uninstall", "--all"}
	out, err := internal.ExecuteCmd(fullPath, args...)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to uninstall gloo",
			zap.Error(err),
			zap.String("out", out))
	}
	return err
}
