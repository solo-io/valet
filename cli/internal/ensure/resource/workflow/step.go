package workflow

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Step struct {
	Curl        *Curl                 `yaml:"curl"`
	Condition   *Condition            `yaml:"condition"`
	DnsEntry    *DnsEntry             `yaml:"dnsEntry"`
	Install     *application.Ref      `yaml:"install"`
	Uninstall   *application.Ref      `yaml:"uninstall"`
	WorkflowRef *Ref                  `yaml:"workflow"`
	Apply       *application.Manifest `yaml:"apply"`
	Delete      *application.Manifest `yaml:"delete"`
	Patch       *Patch                 `yaml:"patch"`

	Values render.Values `yaml:"values"`
	Flags  render.Flags  `yaml:"flags"`
}

func (s *Step) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(s.Values)
	if s.Delete != nil {
		return s.Delete.Teardown(ctx, input, command)
	}
	return resource.EnsureFirst(ctx, input, command, s.Curl, s.Condition, s.DnsEntry, s.Install, s.Uninstall, s.WorkflowRef, s.Patch)
}

func (s *Step) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	input = input.MergeValues(s.Values)
	return resource.TeardownFirst(ctx, input, command, s.Curl, s.Condition, s.DnsEntry, s.Install, s.Uninstall, s.WorkflowRef, s.Patch)
}
