package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Resource interface {
	Ensure(ctx context.Context, inputs InputParams, command cmd.Factory) error
	Teardown(ctx context.Context, inputs InputParams, command cmd.Factory) error
}

type InputParams struct {
	Values Values
	Flags  Flags
	Step   bool
}

func (i *InputParams) DeepCopy() InputParams {
	var flags []string
	flags = append(flags, i.Flags...)
	values := make(map[string]string)
	for k, v := range i.Values {
		values[k] = v
	}
	return InputParams{
		Flags:  flags,
		Values: values,
		Step:   i.Step,
	}
}

func (i *InputParams) MergeValues(values Values) InputParams {
	output := i.DeepCopy()
	for k, v := range values {
		if !output.Values.ContainsKey(k) {
			output.Values[k] = v
		}
	}
	return output
}

func (i *InputParams) MergeFlags(flags Flags) InputParams {
	output := i.DeepCopy()
	for _, flag := range flags {
		found := false
		for _, existingFlag := range output.Flags {
			if flag == existingFlag {
				found = true
				break
			}
		}
		if !found {
			output.Flags = append(output.Flags, flag)
		}
	}
	return output
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

	_ Resource = new(Workflow)
	_ Resource = new(Condition)
	_ Resource = new(Curl)
	_ Resource = new(DnsEntry)
	_ Resource = new(Step)
)

type ApplicationResource struct {
	Namespace   *Namespace      `yaml:"namespace"`
	HelmChart   *HelmChart      `yaml:"helmChart"`
	Secret      *Secret         `yaml:"secret"`
	Path        string          `yaml:"path"`
	Template    *Template       `yaml:"template"`
	Patch       *Patch          `yaml:"patch"`
	Application *ApplicationRef `yaml:"application"`

	Values Values `yaml:"values"`
	Flags  Flags  `yaml:"flags"`
}

func (a *ApplicationResource) Ensure(ctx context.Context, input InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	var manifest *Manifest = nil
	if a.Path != "" {
		manifest = &Manifest{
			a.Path,
		}
	}
	return EnsureFirst(ctx, input, command, a.HelmChart, a.Secret, manifest, a.Template, a.Patch, a.Namespace, a.Application)
}

func (a *ApplicationResource) Teardown(ctx context.Context, input InputParams, command cmd.Factory) error {
	input = input.MergeValues(a.Values)
	var manifest *Manifest = nil
	if a.Path != "" {
		manifest = &Manifest{
			a.Path,
		}
	}
	return TeardownFirst(ctx, input, command, a.HelmChart, a.Secret, manifest, a.Template, a.Patch, a.Namespace, a.Application)
}
