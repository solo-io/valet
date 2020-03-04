package workflow

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

const (
	RenderAsYamlFlag = "RenderAsYaml"
)

type Docs struct {
	Title       string   `json:"title"`
	Description []string `json:"description"`

	DocsValues render.Values `json:"docsValues"`
}

type Section struct {
	Docs
	Sections []Section `json:"sections"`
}

type Documented interface {
	Document(ctx context.Context, input render.InputParams, section *Section) error
}

func DocumentFirst(ctx context.Context, input render.InputParams, section *Section, resources ...Documented) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		return resource.Document(ctx, input, section)
	}
	return nil
}

type Step struct {
	Docs

	Curl        *Curl                 `json:"curl"`
	Condition   *Condition            `json:"condition"`
	DnsEntry    *DnsEntry             `json:"dnsEntry"`
	Install     *application.Ref      `json:"install"`
	Uninstall   *application.Ref      `json:"uninstall"`
	WorkflowRef *Ref                  `json:"workflow"`
	Apply       *application.Resource `json:"apply"`
	Delete      *application.Resource `json:"delete"`
	Patch       *Patch                `json:"patch"`
	Helm3Deploy *Helm3Deploy          `json:"helm3Deploy"`
	RestartPods *RestartPods          `json:"restartPods"`

	Values render.Values `json:"values"`
	Flags  render.Flags  `json:"flags"`
}

func (s *Step) Ensure(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(s.Values)
	if s.Uninstall != nil || s.Delete != nil {
		return resource.TeardownFirst(ctx, input, s.Uninstall, s.Delete)
	}
	return resource.EnsureFirst(ctx, input, s.Curl, s.Condition, s.DnsEntry, s.Install, s.WorkflowRef, s.Apply, s.Patch, s.Helm3Deploy, s.RestartPods)
}

func (s *Step) Teardown(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(s.Values)
	return resource.TeardownFirst(ctx, input, s.Curl, s.Condition, s.DnsEntry, s.Install, s.WorkflowRef, s.Apply, s.Patch, s.Helm3Deploy, s.RestartPods)
}

func (s *Step) Document(ctx context.Context, input render.InputParams, section *Section) error {
	section.Title = s.Title
	section.Description = s.Description
	input = input.MergeValues(s.DocsValues)
	input = input.MergeValues(s.Values)

	if s.Apply != nil {
		if err := s.documentApplyManifests(input, section, s.Apply); err != nil {
			return err
		} else if err := s.documentApplySecret(input, section, s.Apply); err != nil {
			return err
		} else if err := s.documentApplyTemplate(input, section, s.Apply); err != nil {
			return err
		}
	} else {
		return DocumentFirst(ctx, input, section, s.WorkflowRef)
	}
	return nil
}

func (s *Step) documentApplyTemplate(input render.InputParams, section *Section, resource *application.Resource) error {
	if resource.Template == nil {
		return nil
	}

	renderAsYaml := false
	if input.Values.ContainsKey(RenderAsYamlFlag) {
		renderAsYamlVal, err := input.Values.GetValue(RenderAsYamlFlag, input.Runner())
		if err != nil {
			return err
		}
		renderAsYaml = renderAsYamlVal == "true"
	}

	if !renderAsYaml {
		return nil
	}

	template, err := resource.Template.Load(input)
	if err != nil {
		return err
	}

	describe := "```\n" + template + "\n```"
	section.Description = append(section.Description, describe)
	return nil
}

func (s *Step) documentApplySecret(input render.InputParams, section *Section, resource *application.Resource) error {
	if resource.Secret == nil {
		return nil
	}
	full := "kubectl create secret generic"
	if resource.Secret.Namespace != "" {
		full += " -n  " + resource.Secret.Namespace
	}
	full += " " + resource.Secret.Name
	for key, entry := range resource.Secret.Entries {
		if entry.EnvVar != "" {
			full += fmt.Sprintf(" --from-env=%s=$%s", key, entry.EnvVar)
		} else if entry.File != "" {
			full += fmt.Sprintf(" --from-file=%s=%s", key, entry.File)
		}
	}
	describe := "```\n" + full + "\n```"
	section.Description = append(section.Description, describe)
	return nil
}

func (s *Step) documentApplyManifests(input render.InputParams, section *Section, resource *application.Resource) error {
	var manifests []string
	if resource.Manifests != nil {
		manifests = append(manifests, resource.Manifests.Paths...)
	} else if resource.Manifest != nil {
		manifests = append(manifests, resource.Manifest.Path)
	}

	if len(manifests) == 0 {
		return nil
	}

	describe := "```"
	renderAsYaml := false
	if input.Values.ContainsKey(RenderAsYamlFlag) {
		renderAsYamlVal, err := input.Values.GetValue(RenderAsYamlFlag, input.Runner())
		if err != nil {
			return err
		}
		renderAsYaml = renderAsYamlVal == "true"
	}

	if renderAsYaml {
		var yamls []string
		for _, manifest := range manifests {
			yaml, err := input.LoadFile(render.DefaultRegistry, manifest)
			if err != nil {
				return err
			}
			yamls = append(yamls, yaml)
		}
		describe += "yaml\n" + strings.Join(yamls, "\n\n---\n\n")
	} else {
		for _, manifest := range manifests {
			describe += fmt.Sprintf("\nkubectl apply -f %s", manifest)
		}
	}
	describe += "\n```"
	section.Description = append(section.Description, describe)
	return nil
}
