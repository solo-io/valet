package gloo

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/testutils/kube"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func podsReadyAndVersionsMatch(ctx context.Context, namespace, selector, version string) (bool, error) {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return false, err
	}
	pods, err := kubeClient.CoreV1().Pods(namespace).List(v1.ListOptions{LabelSelector: selector})
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error listing pods", zap.Error(err))
		return false, err
	}
	if len(pods.Items) == 0 {
		contextutils.LoggerFrom(ctx).Infow("No pods")
		return false, nil
	}
	for _, pod := range pods.Items {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == v12.ContainersReady && cond.Status != v12.ConditionTrue {
				contextutils.LoggerFrom(ctx).Infow("Pods not ready")
				return false, nil
			}
		}
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if strings.Contains(container.Image, version) {
				return true, nil
			}
		}
	}
	contextutils.LoggerFrom(ctx).Warnw("Detected install, but did not find any containers with the expected version",
		zap.String("expected", version))
	return false, nil
}

func getLatestTag(ctx context.Context, repo string) (string, error) {
	client := githubutils.GetClientWithOrWithoutToken(ctx)
	release, _, err := client.Repositories.GetLatestRelease(ctx, "solo-io", repo)
	if err != nil {
		wrapped := CouldNotDetermineVersionError(err)
		contextutils.LoggerFrom(ctx).Errorw(err.Error(), zap.Error(err))
		return "", wrapped
	}
	return release.GetTagName(), nil
}
