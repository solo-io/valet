package cmd

import "fmt"

type helm Command

func (h *helm) With(args... string) *helm {
	h.Args = append(h.Args, args...)
	return h
}

func (h *helm) Command() Command {
	return Command(*h)
}

func (h *helm) Run() error {
	return h.Command().Run()
}

func (h *helm) Output() (string, error) {
	return h.Command().Output()
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

func Helm(args... string) *helm {
	return &helm{
		Name: "helm",
		Args: args,
	}
}