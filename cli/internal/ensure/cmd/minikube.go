package cmd

import (
	"context"
	"fmt"
)

type minikube Command

func (m *minikube) Delete() *minikube {
	return m.With("delete")
}

func (m *minikube) Status() *minikube {
	return m.With("status")
}

func (m *minikube) Start() *minikube {
	return m.With("start")
}

func (m *minikube) Memory(mb int) *minikube {
	return m.With(fmt.Sprintf("--memory=%d", mb))
}

func (m *minikube) Cpus(cpus int) *minikube {
	return m.With(fmt.Sprintf("--cpus=%d", cpus))
}

func (m *minikube) KubeVersion(kubeVersion string) *minikube {
	return m.With(fmt.Sprintf("--kubernetes-version=%s", kubeVersion))
}

func (m *minikube) With(args ...string) *minikube {
	m.Args = append(m.Args, args...)
	return m
}

func (m *minikube) SwallowError() *minikube {
	m.SwallowErrorLog = true
	return m
}

func (m *minikube) Command() *Command {
	return &Command{
		Name:            m.Name,
		Args:            m.Args,
		StdIn:           m.StdIn,
		Redactions:      m.Redactions,
		SwallowErrorLog: m.SwallowErrorLog,
	}
}

func (m *minikube) Run(ctx context.Context) error {
	return m.Command().Run(ctx)
}

func Minikube(args ...string) *minikube {
	return &minikube{
		Name: "minikube",
		Args: args,
	}
}
