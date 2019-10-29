package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Resource interface {
	Ensure(ctx context.Context, command cmd.Factory) error
	Teardown(ctx context.Context, command cmd.Factory) error
}

var (
	_ Resource = new(Config)

	_ Resource = new(Cluster)
	_ Resource = new(GKE)
	_ Resource = new(Minikube)

	_ Resource = new(Application)
	_ Resource = new(ApplicationResource)
	_ Resource = new(HelmChart)
	_ Resource = new(Manifest)
	_ Resource = new(Secret)
	_ Resource = new(Namespace)
	_ Resource = new(Template)
	_ Resource = new(Patch)
	_ Resource = new(DnsEntry)
	_ Resource = new(Condition)

	_ Resource = new(Workflow)
	_ Resource = new(Curl)
	_ Resource = new(Step)
)

type ApplicationResource struct {
	Namespace *Namespace `yaml:"namespace"`
	HelmChart *HelmChart `yaml:"helmChart"`
	Secret    *Secret    `yaml:"secret"`
	Path      string     `yaml:"path"`
	Template  *Template  `yaml:"template"`
	DnsEntry  *DnsEntry  `yaml:"dnsEntry"`
	Patch     *Patch     `yaml:"patch"`
	Condition *Condition `yaml:"condition"`
	Application *ApplicationRef `yaml:"application"`

	Values    map[string]string `yaml:"values"`
	EnvValues map[string]string `yaml:"envValues"`
}

func (a *ApplicationResource) setValue(key, value string) {
	if a.Values == nil {
		a.Values = make(map[string]string)
	}
	if _, ok := a.Values[key]; !ok {
		a.Values[key] = value
	}
}

func (a *ApplicationResource) setEnvValue(key, value string) {
	if a.EnvValues == nil {
		a.EnvValues = make(map[string]string)
	}
	if _, ok := a.EnvValues[key]; !ok {
		a.EnvValues[key] = value
	}
}

func (a *ApplicationResource) Ensure(ctx context.Context, command cmd.Factory) error {
	if a.HelmChart != nil {
		a.HelmChart.updateWithValues(a.Values)
		return a.HelmChart.Ensure(ctx, command)
	}
	if a.Secret != nil {
		a.Secret.updateWithValues(a.Values)
		return a.Secret.Ensure(ctx, command)
	}
	if a.Path != "" {
		manifest := Manifest{
			Path: a.Path,
		}
		return manifest.Ensure(ctx, command)
	}
	if a.Template != nil {
		mergeValuesForTemplate(a, a.Template)
		return a.Template.Ensure(ctx, command)
	}
	if a.DnsEntry != nil {
		a.DnsEntry.updateWithValues(a.Values)
		return a.DnsEntry.Ensure(ctx, command)
	}
	if a.Patch != nil {
		return a.Patch.Ensure(ctx, command)
	}
	if a.Condition != nil {
		return a.Condition.Ensure(ctx, command)
	}
	if a.Namespace != nil {
		return a.Namespace.Ensure(ctx, command)
	}
	if a.Application != nil {
		a.Application.updateWithValues(a.Values)
		return a.Application.Ensure(ctx, command)
	}
	return nil
}

func mergeValuesForTemplate(resource *ApplicationResource, template *Template) {
	for k, v := range resource.Values {
		template.setValue(k, v)
	}
}

func (a *ApplicationResource) Teardown(ctx context.Context, command cmd.Factory) error {
	if a.HelmChart != nil {
		a.HelmChart.updateWithValues(a.Values)
		return a.HelmChart.Teardown(ctx, command)
	}
	if a.Secret != nil {
		a.Secret.updateWithValues(a.Values)
		return a.Secret.Teardown(ctx, command)
	}
	if a.Path != "" {
		manifest := Manifest{
			Path: a.Path,
		}
		return manifest.Teardown(ctx, command)
	}
	if a.Template != nil {
		mergeValuesForTemplate(a, a.Template)
		return a.Template.Teardown(ctx, command)
	}
	if a.DnsEntry != nil {
		a.DnsEntry.updateWithValues(a.Values)
		return a.DnsEntry.Teardown(ctx, command)
	}
	if a.Patch != nil {
		return a.Patch.Teardown(ctx, command)
	}
	if a.Condition != nil {
		return a.Condition.Teardown(ctx, command)
	}
	if a.Namespace != nil {
		return a.Namespace.Teardown(ctx, command)
	}
	if a.Application != nil {
		a.Application.updateWithValues(a.Values)
		return a.Application.Teardown(ctx, command)
	}
	return nil
}
