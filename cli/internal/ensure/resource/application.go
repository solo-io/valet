package resource

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	_ Resource = new(Application)
)

type Application struct {
	Name          string                `yaml:"name"`
	Version       string                `yaml:"version"`
	Namespace     string                `yaml:"namespace"`
	Resources     []ApplicationResource `yaml:"resources"`
	LabelSelector string                `yaml:"labelSelector"`
	From          *ApplicationRef       `yaml:"from"`

	Values    map[string]string `yaml:"values"`
}

type ApplicationRef struct {
	Path string `yaml:"path"`
}

func (a *Application) Teardown(ctx context.Context, command cmd.Factory) error {
	app, err := a.reconcile(ctx, command)
	if err != nil {
		return err
	}
	var resources []Resource
	for i := len(app.Resources) - 1; i >= 0; i-- {
		r := &app.Resources[i]
		mergeValues(app, r)
		resources = append(resources, r)
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
		mergeValues(app, &r)
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

func mergeValues(app *Application, resource *ApplicationResource) {
	for k, v := range app.Values {
		resource.setValue(k, v)
	}
	if app.Version != "" {
		resource.setValue(VersionKey, app.Version)
	}
	if app.Namespace != "" {
		resource.setValue(NamespaceKey, app.Namespace)
	}
}

func (a *Application) reconcile(ctx context.Context, command cmd.Factory) (*Application, error) {
	if a.From == nil {
		return a, nil
	}
	from, err := LoadApplication(ctx, a.From.Path)
	if err != nil {
		return nil, err
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
	return from, nil
}

func LoadApplication(ctx context.Context, path string) (*Application, error) {
	var a Application

	b, err := loadBytesFromPath(ctx, path)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, &a); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to unmarshal file",
			zap.Error(err),
			zap.String("path", path),
			zap.ByteString("bytes", b))
		return nil, err
	}

	return &a, nil
}
