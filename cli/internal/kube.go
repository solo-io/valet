package internal

import (
	"strings"
	"time"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/testutils/kube"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	v12 "k8s.io/api/core/v1"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	TimedOutWaitingForPodsError = errors.Errorf("Timed out waiting for pods to come online")
)

func WaitUntilPodsRunning(namespace string) error {
	cmd.Stdout().Println("Waiting for pods")
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
			if pod.Status.Phase == v12.PodSucceeded {
				continue
			}
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
			cmd.Stderr().Println("Timed out waiting for pods to come online: %v", notYetRunning)
			return TimedOutWaitingForPodsError
		case <-time.After(time.Second / 2):
			notYetRunning = make(map[string]struct{})
			ready, err := podsReady()
			if err != nil {
				cmd.Stderr().Println("Error checking for ready pods: %s", err.Error())
				return err
			}
			if ready {
				cmd.Stdout().Println("Pods are ready")
				return nil
			}
		}
	}
}

func NamespaceIsActive(namespace string) (bool, error) {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return false, err
	}
	ns, err := kubeClient.CoreV1().Namespaces().Get(namespace, v1.GetOptions{})
	if err != nil {
		if kubeerrs.IsNotFound(err) {
			return false, nil
		}
		cmd.Stderr().Println("Error trying to get namespace %s: %s", namespace, err.Error())
		return false, err
	}
	if ns.Status.Phase != v12.NamespaceActive {
		cmd.Stderr().Println("Namespace is not active (%s)", ns.Status.Phase)
	}
	return true, nil
}

func PodsReadyAndVersionsMatch(namespace, selector, version string) (bool, error) {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return false, err
	}
	pods, err := kubeClient.CoreV1().Pods(namespace).List(v1.ListOptions{LabelSelector: selector})
	if err != nil {
		cmd.Stderr().Println("Error listing pods: %s", err.Error())
		return false, err
	}
	if len(pods.Items) == 0 {
		cmd.Stdout().Println("No pods")
		return false, nil
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == v12.PodSucceeded {
			continue
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == v12.ContainersReady && cond.Status != v12.ConditionTrue {
				cmd.Stdout().Println("Pods not ready")
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
	cmd.Stdout().Println("Detected install, but did not find any containers with the expected version %s", version)
	return false, nil
}
