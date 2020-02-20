package workflow

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

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
