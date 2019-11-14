package cmd

import (
	"context"
	"fmt"
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

func (m *Minikube) Delete() *Minikube {
	return m.With("delete")
}

func (m *Minikube) Status() *Minikube {
	return m.With("status")
}

func (m *Minikube) Start() *Minikube {
	return m.With("start")
}

func (m *Minikube) Memory(mb int) *Minikube {
	return m.With(fmt.Sprintf("--memory=%d", mb))
}

func (m *Minikube) Cpus(cpus int) *Minikube {
	return m.With(fmt.Sprintf("--cpus=%d", cpus))
}

func (m *Minikube) KubeVersion(kubeVersion string) *Minikube {
	return m.With(fmt.Sprintf("--kubernetes-version=%s", kubeVersion))
}

func (m *Minikube) IP(ctx context.Context, runner Runner) (string, error) {
	return runner.Output(ctx, m.With("ip").Cmd())
}
