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

var (
	_ Resource = new(Application)
)

type Application struct {
	Namespace     *Namespace            `yaml:"namespace"`
	Resources     []ApplicationResource `yaml:"resources"`
	LabelSelector string                `yaml:"labelSelector"`
}

type HelmChart struct {
	RepoUrl   string `yaml:"repoUrl"`
	ChartName string `yaml:"chartName"`
	RepoName  string `yaml:"repoName"`
	Version   string `yaml:"version"`
	Namespace string `yaml:"namespace"` // a bit redundant
	Set       []string `yaml:"set"`
}

type ApplicationResource struct {
	HelmChart *HelmChart
	Secret    *Secret
	Manifest  *Manifest
}

var _ Resource = new(ApplicationResource)

func (a *ApplicationResource) Ensure(ctx context.Context, command cmd.Factory) error {
	return EnsureAll(ctx, command, a.HelmChart, a.Secret, a.Manifest)
}

func (a *ApplicationResource) Teardown(ctx context.Context, command cmd.Factory) error {
	return TeardownAll(ctx, command, a.HelmChart, a.Secret, a.Manifest)
}

func (a *Application) Teardown(ctx context.Context, command cmd.Factory) error {
	var resources []Resource
	for i := len(a.Resources)-1; i >= 0; i-- {
		resources = append(resources, &a.Resources[i])
	}
	if err := TeardownAll(ctx, command, resources...); err != nil {
		return err
	}
	if a.Namespace != nil {
		if err := a.Namespace.Teardown(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) Ensure(ctx context.Context, command cmd.Factory) error {
	if a.Namespace != nil {
		if err := a.Namespace.Ensure(ctx, command); err != nil {
			return err
		}
	}
	var resources []Resource
	for _, resource := range a.Resources {
		resources = append(resources, &resource)
	}
	if err := EnsureAll(ctx, command, resources...); err != nil {
		return err
	}
	if a.Namespace == nil {
		return nil
	}
	return internal.WaitUntilPodsRunning(ctx, a.Namespace.Name)
}

func (h *HelmChart) Ensure(ctx context.Context, command cmd.Factory) error {
	contextutils.LoggerFrom(ctx).Infof("preparing to install %s version %s", h.ChartName, h.Version)
	manifest, err := h.renderManifest(ctx, command)
	if err != nil {
		return err
	}
	if err := command.Kubectl().ApplyStdIn(manifest).Cmd().Run(ctx); err != nil {
		return err
	}
	return internal.WaitUntilPodsRunning(ctx, SmMarketplaceNamespace)
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
	contextutils.LoggerFrom(ctx).Infow("rendering and applying manifest for service mesh hub")
	helmCmd := command.Helm().Template().Namespace(h.Namespace)
	for _, set := range h.Set {
		helmCmd = helmCmd.Set(set)
	}
	return helmCmd.Target(chartDir).Cmd().Output(ctx)
}

func (h *HelmChart) Teardown(ctx context.Context, command cmd.Factory) error {
	contextutils.LoggerFrom(ctx).Infof("preparing to uninstall %s version %s", h.ChartName, h.Version)
	manifest, err := h.renderManifest(ctx, command)
	if err != nil {
		return err
	}
	return command.Kubectl().DeleteStdIn(manifest).Cmd().Run(ctx)
}

func (h *HelmChart) renderAndDeleteManifest(ctx context.Context, command cmd.Factory, chartDir string) error {
	contextutils.LoggerFrom(ctx).Infow("rendering and deleting manifest for service mesh hub")
	out, err := command.Helm().Template().Namespace(h.Namespace).Set("namespace.create=true").Target(chartDir).Cmd().Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error rendering manifest", zap.Error(err), zap.String("out", out))
		return err
	}
	out, err = command.Kubectl().DeleteStdIn(out).Cmd().Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error deleting manifest", zap.Error(err), zap.String("out", out))
	}
	return err
}

func (h *HelmChart) addHelmRepo(ctx context.Context, command cmd.Factory) error {
	return command.Helm().AddRepo(h.RepoName, h.RepoUrl).Cmd().Run(ctx)
}

func (h *HelmChart) fetchAndUntarChart(ctx context.Context, command cmd.Factory) (string, error) {
	untarDir, err := getLocalDirectory(h.Version)
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
