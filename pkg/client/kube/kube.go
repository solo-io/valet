package kube

import (
	"bytes"
	"fmt"
	"github.com/solo-io/go-utils/kubeutils"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/exec"
	"strings"
	"time"

	errors "github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils/kube"
	"github.com/solo-io/valet/pkg/cmd"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -destination ./mocks/kube_client_mock.go github.com/solo-io/valet/pkg/client/kube Client

const (
	LocalClusterName = "minikube"
)

// A simple client for Valet's interactions with Kubernetes via client API
type Client interface {
	// Wait until all of the pods in the provided namespace are ready (or completed successfully)
	WaitUntilPodsRunning(namespace string) error
	// Get the address of the service, trying to account for different service types (i.e. LoadBalancer) and
	// Kubernetes flavors (i.e. Minikube)
	GetIngressAddress(name, namespace, proxyPort string) (string, error)
}

// Create a default kube client
func NewClient() *kubeClient {
	return &kubeClient{}
}

type kubeClient struct {
}

var (
	TimedOutWaitingForPodsError = errors.Errorf("Timed out waiting for pods to come online")
)

func (k *kubeClient) GetIngressAddress(name, namespace, proxyPort string) (string, error) {
	restCfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return "", errors.Wrapf(err, "getting kube rest config")
	}
	kube, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return "", errors.Wrapf(err, "starting kube client")
	}
	svc, err := kube.CoreV1().Services(namespace).Get(name, v12.GetOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "could not detect '%v' service in %v namespace", name, namespace)
	}
	var svcPort *v1.ServicePort
	switch len(svc.Spec.Ports) {
	case 0:
		return "", errors.Errorf("service %v is missing ports", name)
	case 1:
		svcPort = &svc.Spec.Ports[0]
	default:
		for _, p := range svc.Spec.Ports {
			if p.Name == proxyPort {
				svcPort = &p
				break
			}
		}
		if svcPort == nil {
			return "", errors.Errorf("named port %v not found on service %v", proxyPort, name)
		}
	}

	var host, port string
	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		// assume nodeport on kubernetes
		// TODO: support more types of NodePort services
		host, err = getNodeIp(svc, kube)
		if err != nil {
			return "", errors.Wrapf(err, "")
		}
		port = fmt.Sprintf("%v", svcPort.NodePort)
	} else {
		host = svc.Status.LoadBalancer.Ingress[0].Hostname
		if host == "" {
			host = svc.Status.LoadBalancer.Ingress[0].IP
		}
		port = fmt.Sprintf("%v", svcPort.Port)
	}
	return host + ":" + port, nil
}

func getNodeIp(svc *v1.Service, kube kubernetes.Interface) (string, error) {
	// pick a node where one of our pods is running
	pods, err := kube.CoreV1().Pods(svc.Namespace).List(v12.ListOptions{
		LabelSelector: labels.SelectorFromSet(svc.Spec.Selector).String(),
	})
	if err != nil {
		return "", err
	}
	var nodeName string
	for _, pod := range pods.Items {
		if pod.Spec.NodeName != "" {
			nodeName = pod.Spec.NodeName
			break
		}
	}
	if nodeName == "" {
		return "", errors.Errorf("no node found for %v's pods. ensure at least one pod has been deployed "+
			"for the %v service", svc.Name, svc.Name)
	}
	// special case for minikube
	// we run `minikube ip` which avoids an issue where
	// we get a NAT network IP when the minikube provider is virtualbox
	if nodeName == "minikube" {
		return minikubeIp(LocalClusterName)
	}

	node, err := kube.CoreV1().Nodes().Get(nodeName, v12.GetOptions{})
	if err != nil {
		return "", err
	}

	for _, addr := range node.Status.Addresses {
		return addr.Address, nil
	}

	return "", errors.Errorf("no active addresses found for node %v", node.Name)
}

func minikubeIp(clusterName string) (string, error) {
	minikubeCmd := exec.Command("minikube", "ip", "-p", clusterName)

	hostname := &bytes.Buffer{}

	minikubeCmd.Stdout = hostname
	minikubeCmd.Stderr = os.Stderr
	err := minikubeCmd.Run()

	return strings.TrimSuffix(hostname.String(), "\n"), err
}


func (k *kubeClient) WaitUntilPodsRunning(namespace string) error {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return err
	}
	pods := kubeClient.CoreV1().Pods(namespace)
	podsReady := func() (bool, error) {
		list, err := pods.List(v12.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, pod := range list.Items {
			if pod.Status.Phase == v1.PodSucceeded {
				continue
			}
			var podReady bool
			for _, cond := range pod.Status.Conditions {
				if cond.Type == v1.ContainersReady && cond.Status == v1.ConditionTrue {
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

func (k *kubeClient) NamespaceIsActive(namespace string) (bool, error) {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return false, err
	}
	ns, err := kubeClient.CoreV1().Namespaces().Get(namespace, v12.GetOptions{})
	if err != nil {
		if kubeerrs.IsNotFound(err) {
			return false, nil
		}
		cmd.Stderr().Println("Error trying to get namespace %s: %s", namespace, err.Error())
		return false, err
	}
	if ns.Status.Phase != v1.NamespaceActive {
		cmd.Stderr().Println("Namespace is not active (%s)", ns.Status.Phase)
	}
	return true, nil
}

func (k *kubeClient) PodsReadyAndVersionsMatch(namespace, selector, version string) (bool, error) {
	kubeClient, err := kube.KubeClient()
	if err != nil {
		return false, err
	}
	pods, err := kubeClient.CoreV1().Pods(namespace).List(v12.ListOptions{LabelSelector: selector})
	if err != nil {
		cmd.Stderr().Println("Error listing pods: %s", err.Error())
		return false, err
	}
	if len(pods.Items) == 0 {
		cmd.Stdout().Println("No pods")
		return false, nil
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == v1.ContainersReady && cond.Status != v1.ConditionTrue {
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

