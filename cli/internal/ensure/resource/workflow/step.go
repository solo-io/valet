package workflow

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

type Docs struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Notes       string `yaml:"notes"`
}

type Section struct {
	Docs
	Sections []Section `yaml:"sections"`
}

type Documented interface {
	Document(ctx context.Context, input render.InputParams, section *Section)
}

func DocumentFirst(ctx context.Context, input render.InputParams, section *Section, resources ...Documented) {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		resource.Document(ctx, input, section)
		return
	}
}

type Step struct {
	Curl        *Curl                 `yaml:"curl"`
	Condition   *Condition            `yaml:"condition"`
	DnsEntry    *DnsEntry             `yaml:"dnsEntry"`
	Install     *application.Ref      `yaml:"install"`
	Uninstall   *application.Ref      `yaml:"uninstall"`
	WorkflowRef *Ref                  `yaml:"workflow"`
	Apply       *application.Resource `yaml:"apply"`
	Delete      *application.Resource `yaml:"delete"`
	Patch       *Patch                `yaml:"patch"`
	Helm3Deploy *Helm3Deploy          `yaml:"helm3Deploy"`

	Values render.Values `yaml:"values"`
	Flags  render.Flags  `yaml:"flags"`

	Docs         *Docs `yaml:"docs"`
	RenderAsYaml bool  `yaml:"renderAsYaml"`
}

func (s *Step) Ensure(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(s.Values)
	if s.Uninstall != nil || s.Delete != nil {
		return resource.TeardownFirst(ctx, input, s.Uninstall, s.Delete)
	}
	return resource.EnsureFirst(ctx, input, s.Curl, s.Condition, s.DnsEntry, s.Install, s.WorkflowRef, s.Apply, s.Patch, s.Helm3Deploy)
}

func (s *Step) Teardown(ctx context.Context, input render.InputParams) error {
	input = input.MergeValues(s.Values)
	return resource.TeardownFirst(ctx, input, s.Curl, s.Condition, s.DnsEntry, s.Install, s.WorkflowRef, s.Apply, s.Patch, s.Helm3Deploy)
}

func (s *Step) Document(ctx context.Context, input render.InputParams, section *Section) {
	if s.Docs != nil {
		section.Title = s.Docs.Title
		section.Description = s.Docs.Description
		section.Notes = s.Docs.Notes
	}
	if s.Apply != nil {
		if err := s.documentApplyManifests(input, section, s.Apply); err != nil {
			log.Fatal(err)
		} else if err := s.documentApplySecret(input, section, s.Apply); err != nil {
			log.Fatal(err)
		}
	} else {
		DocumentFirst(ctx, input, section, s.WorkflowRef)
	}
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
	section.Description = section.Description + "\n\n" + describe
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
	if s.RenderAsYaml {
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
	section.Description = section.Description + "\n\n" + describe
	return nil
}


