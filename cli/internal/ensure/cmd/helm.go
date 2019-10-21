package cmd

import (
	"fmt"
	"os"
)

type Helm struct {
	cmd *Command
}

func (h *Helm) With(args... string) *Helm {
	h.cmd = h.cmd.With(args...)
	return h
}

func (h *Helm) Cmd() *Command {
	return h.cmd
}

func (h *Helm) AddRepo(repoName, repoUrl string) *Helm {
	return h.With("repo", "add", repoName, repoUrl)
}

func (h *Helm) Template() *Helm {
	return h.With("template")
}

func (h *Helm) Namespace(namespace string) *Helm {
	if namespace == "" {
		return h
	}
	return h.With("--namespace", namespace)
}

func (h *Helm) Set(set string) *Helm {
	return h.With("--set", set)
}

func (h *Helm) SetEnv(set, envVar string) *Helm {
	envVarValue := os.Getenv(envVar)
	unredacted := fmt.Sprintf("%s=%s", set, envVarValue)
	redacted := fmt.Sprintf("%s=%s", set, Redacted)
	h.cmd = h.cmd.Redact(unredacted, redacted)
	return h.With("--set", )
}

func (h *Helm) Target(target string) *Helm {
	return h.With(target)
}

func (h *Helm) Fetch(repoName, chartName string) *Helm {
	return h.With("fetch", fmt.Sprintf("%s/%s", repoName, chartName))
}

func (h *Helm) Version(version string) *Helm {
	return h.With("--version", version)
}

func (h *Helm) UntarToDir(untarDir string) *Helm {
	return h.With("--untar", "--untardir", untarDir)
}
