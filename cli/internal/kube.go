package internal

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/testutils/kube"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func WaitUntilPodsRunning(ctx context.Context, namespace string) error {
	contextutils.LoggerFrom(ctx).Infow("Waiting for pods")
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return err
	}
	pods := kubeClient.CoreV1().Pods(namespace)
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
				contextutils.LoggerFrom(ctx).Infow("pods are ready")
				return nil
			}
		}
	}
}

func NamespaceIsActive(ctx context.Context, namespace string) (bool, error) {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return false, err
	}
	ns, err := kubeClient.CoreV1().Namespaces().Get(namespace, v1.GetOptions{})
	if err != nil {
		if kubeerrs.IsNotFound(err) {
			return false, nil
		}
		contextutils.LoggerFrom(ctx).Errorw("Error trying to get namespace", zap.Error(err), zap.String("ns", namespace))
		return false, err
	}
	if ns.Status.Phase != v12.NamespaceActive {
		contextutils.LoggerFrom(ctx).Errorw("Namespace is not active", zap.Any("phase", ns.Status.Phase))
	}
	return true, nil
}
