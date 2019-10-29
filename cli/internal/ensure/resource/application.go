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
	RequiredValueNotProvidedError = func(key string) error {
		return errors.Errorf("Required value %s not found", key)
	}
)

type Application struct {
	Name          string                `yaml:"name"`
	Version       string                `yaml:"version"`
	Resources     []ApplicationResource `yaml:"resources"`
	LabelSelector string                `yaml:"labelSelector"`

	RequiredValues []string          `yaml:"requiredValues"`
	Values         map[string]string `yaml:"values"`
}

type ApplicationRef struct {
	Path   string            `yaml:"path"`
	Values map[string]string `yaml:"values"`
}

func (a *ApplicationRef) updateWithValues(values map[string]string) {
	for k, v := range values {
		if a.Values == nil {
			a.Values = make(map[string]string)
		}
		a.Values[k] = v
	}
}

func (a *ApplicationRef) load(ctx context.Context) (*Application, error) {
	app, err := LoadApplication(ctx, a.Path)
	if err != nil {
		return nil, err
	}
	for k, v := range a.Values {
		if app.Values == nil {
			app.Values = make(map[string]string)
		}
		app.Values[k] = v
	}
	return app, nil
}

func (a *ApplicationRef) Ensure(ctx context.Context, command cmd.Factory) error {
	app, err := a.load(ctx)
	if err != nil {
		return err
	}
	return app.Ensure(ctx, command)
}

func (a *ApplicationRef) Teardown(ctx context.Context, command cmd.Factory) error {
	app, err := a.load(ctx)
	if err != nil {
		return err
	}
	return app.Teardown(ctx, command)
}

func (a *Application) checkRequiredValues() error {
	for _, key := range a.RequiredValues {
		if a.Values == nil {
			return RequiredValueNotProvidedError(key)
		}
		if val, ok := a.Values[key]; !ok || val == "" {
			return RequiredValueNotProvidedError(key)
		}
	}
	return nil
}

func (a *Application) Teardown(ctx context.Context, command cmd.Factory) error {
	if err := a.checkRequiredValues(); err != nil {
		return err
	}
	var resources []Resource
	for i := len(a.Resources) - 1; i >= 0; i-- {
		r := &a.Resources[i]
		mergeValues(a, r)
		resources = append(resources, r)
	}
	if err := TeardownAll(ctx, command, resources...); err != nil {
		return err
	}
	if a.Values != nil {
		if val, ok := a.Values[NamespaceKey]; ok {
			namespace := Namespace{
				Name: val,
			}
			if err := namespace.Teardown(ctx, command); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Application) Ensure(ctx context.Context, command cmd.Factory) error {
	if err := a.checkRequiredValues(); err != nil {
		return err
	}
	ensuredNamespace := ""
	if a.Values != nil {
		if val, ok := a.Values[NamespaceKey]; ok {
			namespace := Namespace{
				Name: val,
			}
			if err := namespace.Ensure(ctx, command); err != nil {
				return err
			}
			ensuredNamespace = val
		}
	}
	var resources []Resource
	for _, resource := range a.Resources {
		r := resource
		mergeValues(a, &r)
		resources = append(resources, &r)
	}
	if err := EnsureAll(ctx, command, resources...); err != nil {
		return err
	}
	if ensuredNamespace == "" {
		return nil
	}
	return internal.WaitUntilPodsRunning(ctx, ensuredNamespace)
}

func mergeValues(app *Application, resource *ApplicationResource) {
	for k, v := range app.Values {
		resource.setValue(k, v)
	}
	if app.Version != "" {
		resource.setValue(VersionKey, app.Version)
	}
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
