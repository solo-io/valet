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

	Values Values   `yaml:"values"`
	Flags  []string `yaml:"flags"`
}

func (s *Step) updateWithValues(values Values) {
	s.Values = MergeValues(values, s.Values)
}

func (s *Step) updateWithFlags(flags []string) {
	s.Flags = append(s.Flags, flags...)
}

func (s *Step) Ensure(ctx context.Context, command cmd.Factory) error {
	if s.Install != nil {
		s.Install.updateWithValues(s.Values)
		s.Install.updateWithFlags(s.Flags)
		if err := s.Install.Ensure(ctx, command); err != nil {
			return err
		}
	}
	if s.Uninstall != nil {
		s.Uninstall.updateWithFlags(s.Flags)
		s.Uninstall.updateWithValues(s.Values)
		if err := s.Uninstall.Teardown(ctx, command); err != nil {
			return err
		}
	}
	if s.Curl != nil {
		if err := s.Curl.Ensure(ctx, command); err != nil {
			return err
		}
	}
	if s.Condition != nil {
		if err := s.Condition.Ensure(ctx, command); err != nil {
			return err
		}
	}
	if s.DnsEntry != nil {
		if err := s.DnsEntry.updateWithValues(s.Values); err != nil {
			return err
		}
		if err := s.DnsEntry.Ensure(ctx, command); err != nil {
			return err
		}
	}
	if s.WorkflowRef != nil {
		s.WorkflowRef.updateWithValues(s.Values)
		s.WorkflowRef.updateWithFlags(s.Flags)
		if err := s.WorkflowRef.Ensure(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (s *Step) Teardown(ctx context.Context, command cmd.Factory) error {
	if s.DnsEntry != nil {
		if err := s.DnsEntry.Teardown(ctx, command); err != nil {
			return err
		}
	}
	// TODO: Figure out more teardown story for workflows
	return nil
}
