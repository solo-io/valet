package resource

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

type HelmChart struct {
	RepoUrl   string            `yaml:"repoUrl"`
	ChartName string            `yaml:"chartName"`
	RepoName  string            `yaml:"repoName"`
	Version   string            `yaml:"version"`
	Namespace string            `yaml:"namespace"` // a bit redundant
	Set       []string          `yaml:"set"`
	SetEnv    map[string]string `yaml:"setEnv"`
}

func (h *HelmChart) updateWithValues(values map[string]string) {
	if h.Version == "" {
		if val, ok := values[VersionKey]; ok {
			h.Version = val
		}
	}
	if h.Namespace == "" {
		if val, ok := values[NamespaceKey]; ok {
			h.Namespace = val
		}
	}
}

func (h *HelmChart) Ensure(ctx context.Context, command cmd.Factory) error {
	contextutils.LoggerFrom(ctx).Infof("preparing to install %s version %s", h.ChartName, h.Version)
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
	return internal.WaitUntilPodsRunning(ctx, h.Namespace)
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
	contextutils.LoggerFrom(ctx).Infow("rendering and applying manifest for application")
	helmCmd := command.Helm().Template().Namespace(h.Namespace)
	for _, set := range h.Set {
		helmCmd = helmCmd.Set(set)
	}
	for set, envVar := range h.SetEnv {
		helmCmd = helmCmd.SetEnv(set, envVar)
	}
	return helmCmd.Target(chartDir).Cmd().Output(ctx)
}

func (h *HelmChart) Teardown(ctx context.Context, command cmd.Factory) error {
	contextutils.LoggerFrom(ctx).Infof("preparing to uninstall %s version %s", h.ChartName, h.Version)
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
		if ok, err := internal.NamespaceIsActive(ctx, h.Namespace); err != nil {
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
		contextutils.LoggerFrom(ctx).Errorw("error determining local directory for untarring chart", zap.Error(err))
		return "", err
	}
	if err := os.MkdirAll(untarDir, os.ModePerm); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error making directory", zap.Error(err))
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
		contextutils.LoggerFrom(ctx).Errorw("error trying to untar helm chart", zap.Error(err), zap.String("out", out))
		return "", err
	}
	contextutils.LoggerFrom(ctx).Infow("successfully downloaded and extracted chart", zap.String("untarDir", untarDir))
	return untarDir, nil
}

func (h *HelmChart) getLocalDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".helm", "untar", h.RepoName, h.ChartName, h.Version), nil
}
