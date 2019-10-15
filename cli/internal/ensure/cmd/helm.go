package cmd

import (
	"context"
	"fmt"
)

type helm Command

func (h *helm) With(args ...string) *helm {
	h.Args = append(h.Args, args...)
	return h
}

func (h *helm) Command() *Command {
	return &Command{
		Name:       h.Name,
		Args:       h.Args,
		StdIn:      h.StdIn,
		Redactions: h.Redactions,
	}
}

func (h *helm) Run(ctx context.Context) error {
	return h.Command().Run(ctx)
}

func (h *helm) Output(ctx context.Context) (string, error) {
	return h.Command().Output(ctx)
}

func (h *helm) AddRepo(repoName, repoUrl string) *helm {
	return h.With("repo", "add", repoName, repoUrl)
}

func (h *helm) Template() *helm {
	return h.With("template")
}

func (h *helm) Namespace(namespace string) *helm {
	return h.With("--namespace", namespace)
}

func (h *helm) Set(set string) *helm {
	return h.With("--set", set)
}

func (h *helm) Target(target string) *helm {
	return h.With(target)
}

func (h *helm) Fetch(repoName, chartName string) *helm {
	return h.With("fetch", fmt.Sprintf("%s/%s", repoName, chartName))
}

func (h *helm) Version(version string) *helm {
	return h.With("--version", version)
}

func (h *helm) UntarToDir(untarDir string) *helm {
	return h.With("--untar", "--untardir", untarDir)
}

func Helm(args ...string) *helm {
	return &helm{
		Name: "helm",
		Args: args,
	}
}
