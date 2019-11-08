package resource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type HelmChart struct {
	RepoUrl   string            `yaml:"repoUrl"`
	ChartName string            `yaml:"chartName"`
	RepoName  string            `yaml:"repoName"`
	Version   string            `yaml:"version"`
	Namespace string            `yaml:"namespace"` // a bit redundant
	Set       []string          `yaml:"set"`
	SetEnv    map[string]string `yaml:"setEnv"`
	Values    Values            `yaml:"values"`
	Files     Values            `yaml:"files"`
}

func (h *HelmChart) updateWithValues(values Values) error {
	if h.Version == "" {
		if values.ContainsKey(VersionKey) {
			if val, err := values.GetValue(VersionKey); err != nil {
				return err
			} else {
				h.Version = val
			}
		}
	}
	if h.Namespace == "" {
		if values.ContainsKey(NamespaceKey) {
			if val, err := values.GetValue(NamespaceKey); err != nil {
				return err
			} else {
				h.Namespace = val
			}
		}
	}
	h.Values = MergeValues(h.Values, values)
	return nil
}

func (h *HelmChart) Ensure(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Preparing to install %s version %s", h.ChartName, h.Version)
	manifest, err := h.renderManifest(ctx, command)
	if err != nil {
		return err
	}
	kubectl := command.Kubectl().ApplyStdIn(manifest)
	if h.Namespace != "" {
		kubectl = kubectl.Namespace(h.Namespace)
	}
	if err := kubectl.Cmd().Run(ctx); err != nil {
		return err
	}
	return internal.WaitUntilPodsRunning(h.Namespace)
}

func (h *HelmChart) renderManifest(ctx context.Context, command cmd.Factory) (string, error) {
	if err := h.addHelmRepo(ctx, command); err != nil {
		return "", err
	}
	untarDir, err := h.fetchAndUntarChart(ctx, command)
	chartDir := filepath.Join(untarDir, h.ChartName)
	if err != nil {
		return "", err
	}
	cmd.Stdout().Println("Rendering and applying manifest for application")
	helmCmd := command.Helm().Template().Namespace(h.Namespace)
	for _, set := range h.Set {
		helmCmd = helmCmd.Set(set)
	}
	for set, envVar := range h.SetEnv {
		helmCmd = helmCmd.SetEnv(set, envVar)
	}
	vals, err := renderStringValues(h.Values)
	if err != nil {
		return "", err
	}
	for key, val := range vals {
		helmCmd = helmCmd.Set(fmt.Sprintf("%s=%s", key, val))
	}
	files, err := renderStringValues(h.Files)
	if err != nil {
		return"", err
	}
	for key, file := range files {
		helmCmd = helmCmd.SetFile(fmt.Sprintf("%s=%s", key, file))
	}
	return helmCmd.Target(chartDir).Cmd().Output(ctx)
}

func (h *HelmChart) Teardown(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Preparing to uninstall %s version %s", h.ChartName, h.Version)
	manifest, err := h.renderManifest(ctx, command)
	if err != nil {
		return err
	}
	kubectl := command.Kubectl().DeleteStdIn(manifest).IgnoreNotFound()
	if h.Namespace != "" {
		kubectl = kubectl.Namespace(h.Namespace)
	}
	if err := kubectl.Cmd().Run(ctx); err != nil {
		return err
	}
	if h.Namespace != "" {
		if ok, err := internal.NamespaceIsActive(h.Namespace); err != nil {
			return err
		} else if ok {
			return command.Kubectl().Delete(ns).WithName(h.Namespace).Cmd().Run(ctx)
		}
	}
	return nil
}

func (h *HelmChart) addHelmRepo(ctx context.Context, command cmd.Factory) error {
	return command.Helm().AddRepo(h.RepoName, h.RepoUrl).Cmd().Run(ctx)
}

func (h *HelmChart) fetchAndUntarChart(ctx context.Context, command cmd.Factory) (string, error) {
	untarDir, err := h.getLocalDirectory()
	if err != nil {
		cmd.Stderr().Println("Error determining local directory for extracting chart: %s", err.Error())
		return "", err
	}
	if err := os.MkdirAll(untarDir, os.ModePerm); err != nil {
		cmd.Stderr().Println("Error making directory: %s", err.Error())
		return "", err
	}
	out, err := command.
		Helm().
		Fetch(h.RepoName, h.ChartName).
		Version(h.Version).
		UntarToDir(untarDir).
		Cmd().
		Output(ctx)
	if err != nil {
		cmd.Stderr().Println("Error trying to extract chart: %s", err.Error())
		cmd.Stderr().Println(out)
		return "", err
	}
	cmd.Stdout().Println("Successfully downloaded and extracted chart")
	return untarDir, nil
}

func (h *HelmChart) getLocalDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".helm", "untar", h.RepoName, h.ChartName, h.Version), nil
}
