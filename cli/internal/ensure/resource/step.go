package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Step struct {
	Curl        *Curl           `yaml:"curl"`
	Condition   *Condition      `yaml:"condition"`
	DnsEntry    *DnsEntry       `yaml:"dnsEntry"`
	Install     *ApplicationRef `yaml:"install"`
	Uninstall   *ApplicationRef `yaml:"uninstall"`
	WorkflowRef *WorkflowRef    `yaml:"workflow"`
	Apply       *Manifest       `yaml:"apply"`
	Delete      *Manifest       `yaml:"delete"`

	Values Values `yaml:"values"`
	Flags  Flags  `yaml:"flags"`
}

func (s *Step) Ensure(ctx context.Context, input InputParams, command cmd.Factory) error {
	input = input.MergeValues(s.Values)
	if s.Delete != nil {
		return s.Delete.Teardown(ctx, input, command)
	}
	return EnsureFirst(ctx, input, command, s.Curl, s.Condition, s.DnsEntry, s.Install, s.Uninstall, s.WorkflowRef, s.Apply)
}

func (s *Step) Teardown(ctx context.Context, input InputParams, command cmd.Factory) error {
	input = input.MergeValues(s.Values)
	return TeardownFirst(ctx, input, command, s.Condition, s.DnsEntry, s.Install, s.Uninstall, s.WorkflowRef, s.Apply)
}
