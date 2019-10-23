package resource

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
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
	Name          string                `yaml:"name"`
	Namespace     string                `yaml:"namespace"`
	Resources     []ApplicationResource `yaml:"resources"`
	LabelSelector string                `yaml:"labelSelector"`
	From          *ApplicationRef       `yaml:"from"`
}

type ApplicationRef struct {
	Path    string                `yaml:"path"`
	Name    string                `yaml:"name"`
	// When loading an application, patches defines a set of resources to edit
	Patches []ApplicationResource `yaml:"patches"`
}

type HelmChart struct {
	RepoUrl   string            `yaml:"repoUrl"`
	ChartName string            `yaml:"chartName"`
	RepoName  string            `yaml:"repoName"`
	Version   string            `yaml:"version"`
	Namespace string            `yaml:"namespace"` // a bit redundant
	Set       []string          `yaml:"set"`
	SetEnv    map[string]string `yaml:"setEnv"`
}

type ApplicationResource struct {
	// Optional: used to look up a resource when trying to patch
	Id        string     `yaml:"id"`
	HelmChart *HelmChart `yaml:"helmChart"`
	Secret    *Secret    `yaml:"secret"`
	Path      string     `yaml:"path"`
	Template  *Template  `yaml:"template"`
	DnsEntry  *DnsEntry  `yaml:"dnsEntry"`
	Cli       *Cli       `yaml:"cli"`
}

type Cli struct {
	Glooctl *Glooctl `yaml:"glooctl"`
}

func (c *Cli) Ensure(ctx context.Context, command cmd.Factory) error {
	if c.Glooctl != nil {
		return c.Glooctl.Ensure(ctx, command)
	}
	return nil
}

func (c *Cli) Teardown(ctx context.Context, command cmd.Factory) error {
	if c.Glooctl != nil {
		return c.Glooctl.Teardown(ctx, command)
	}
	return nil
}

var _ Resource = new(ApplicationResource)

func (a *ApplicationResource) Ensure(ctx context.Context, command cmd.Factory) error {
	if a.HelmChart != nil {
		return a.HelmChart.Ensure(ctx, command)
	}
	if a.Secret != nil {
		return a.Secret.Ensure(ctx, command)
	}
	if a.Path != "" {
		manifest := Manifest{
			Path: a.Path,
		}
		return manifest.Ensure(ctx, command)
	}
	if a.Template != nil {
		return a.Template.Ensure(ctx, command)
	}
	if a.DnsEntry != nil {
		return a.DnsEntry.Ensure(ctx, command)
	}
	return nil
}

func (a *ApplicationResource) Teardown(ctx context.Context, command cmd.Factory) error {
	if a.HelmChart != nil {
		return a.HelmChart.Teardown(ctx, command)
	}
	if a.Secret != nil {
		return a.Secret.Teardown(ctx, command)
	}
	if a.Path != "" {
		manifest := Manifest{
			Path: a.Path,
		}
		return manifest.Teardown(ctx, command)
	}
	if a.Template != nil {
		return a.Template.Teardown(ctx, command)
	}
	if a.DnsEntry != nil {
		return a.DnsEntry.Teardown(ctx, command)
	}
	return nil
}

func (a *Application) Teardown(ctx context.Context, command cmd.Factory) error {
	app, err := a.reconcile(ctx, command)
	if err != nil {
		return err
	}
	var resources []Resource
	for i := len(app.Resources) - 1; i >= 0; i-- {
		resources = append(resources, &app.Resources[i])
	}
	if err := TeardownAll(ctx, command, resources...); err != nil {
		return err
	}
	if app.Namespace != "" {
		namespace := Namespace{
			Name: app.Namespace,
		}
		if err := namespace.Teardown(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) Ensure(ctx context.Context, command cmd.Factory) error {
	app, err := a.reconcile(ctx, command)
	if err != nil {
		return err
	}
	if app.Namespace != "" {
		namespace := Namespace{
			Name: app.Namespace,
		}
		if err := namespace.Ensure(ctx, command); err != nil {
			return err
		}
	}
	var resources []Resource
	for _, resource := range app.Resources {
		r := resource
		resources = append(resources, &r)
	}
	if err := EnsureAll(ctx, command, resources...); err != nil {
		return err
	}
	if app.Namespace == "" {
		return nil
	}
	return internal.WaitUntilPodsRunning(ctx, app.Namespace)
}

func (a *Application) reconcile(ctx context.Context, command cmd.Factory) (*Application, error) {
	if a.From == nil {
		return a, nil
	}
	fromConfig, err := LoadConfig(ctx, a.From.Path)
	if err != nil {
		return nil, err
	}
	var from *Application
	for _, application := range fromConfig.Applications {
		if application.Name == a.From.Name {
			from = &application
			break
		}
	}
	if from == nil {
		return nil, errors.Errorf("Could not find parent")
	}

	if from.From != nil {
		return nil, errors.Errorf("Recursive chaining not yet implemented")
	}
	if a.Namespace != "" {
		from.Namespace = a.Namespace
	}
	from.Resources = append(from.Resources, a.Resources...)
	if a.LabelSelector != "" {
		from.LabelSelector = a.LabelSelector
	}
	for _, patch := range a.From.Patches {
		toPatch := patch.Id
		if toPatch == "" {
			return nil, errors.Errorf("Can't patch an empty id")
		}
		for _, resource := range from.Resources {
			if resource.Id != toPatch {
				continue
			}
			if resource.HelmChart != nil && patch.HelmChart != nil {
				if patch.HelmChart.Namespace != "" {
					resource.HelmChart.Namespace = patch.HelmChart.Namespace
				}
				if patch.HelmChart.Version != "" {
					resource.HelmChart.Version = patch.HelmChart.Version
				}
				if patch.HelmChart.Set != nil {
					resource.HelmChart.Set = append(patch.HelmChart.Set, resource.HelmChart.Set...)
				}
			}
		}
	}
	return from, nil
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



