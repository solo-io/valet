package cmd

import (
	"fmt"
	"strings"

	errors "github.com/rotisserie/eris"
)

type Minikube struct {
	cmd *Command
}

func (m *Minikube) With(args ...string) *Minikube {
	m.cmd = m.cmd.With(args...)
	return m
}

func (m *Minikube) SwallowError() *Minikube {
	m.cmd.SwallowErrorLog = true
	return m
}

func (m *Minikube) Cmd() *Command {
	return m.cmd
}

func (m *Minikube) Status() *Minikube {
	return m.With("status")
}

func (m *Minikube) Memory(mb int) *Minikube {
	return m.With(fmt.Sprintf("--memory=%d", mb))
}

func (m *Minikube) VmDriver(vmDriver string) *Minikube {
	if vmDriver == "" {
		return m
	}
	return m.With(fmt.Sprintf("--vm-driver=%s", vmDriver))
}

func (m *Minikube) Cpus(cpus int) *Minikube {
	return m.With(fmt.Sprintf("--cpus=%d", cpus))
}

func (m *Minikube) KubeVersion(kubeVersion string) *Minikube {
	return m.With(fmt.Sprintf("--kubernetes-version=%s", kubeVersion))
}

func (m *Minikube) FeatureGates(featureGates []string) *Minikube {
	if len(featureGates) == 0 {
		return m
	}
	return m.With(fmt.Sprintf("--feature-gates='%s'", strings.Join(featureGates, ",")))
}

func (m *Minikube) IP(runner Runner) (string, error) {
	return runner.Output(m.With("ip").Cmd())
}

func (m *Minikube) Start(runner Runner) error {
	streamHandler, err := runner.Stream(m.With("start").Cmd())
	if err != nil {
		return err
	}
	inputErr := errors.New("could not delete minikube cluster")
	return streamHandler.StreamHelper(inputErr)
}

func (m *Minikube) Delete(runner Runner) error {
	streamHandler, err := runner.Stream(m.With("delete").Cmd())
	if err != nil {
		return err
	}
	inputErr := errors.New("could not delete minikube cluster")
	return streamHandler.StreamHelper(inputErr)
}
