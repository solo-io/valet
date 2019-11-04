package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Step struct {
	Apply     string     `yaml:"apply"`
	Delete    string     `yaml:"delete"`
	Curl      *Curl      `yaml:"curl"`
	Condition *Condition `yaml:"condition"`
	DnsEntry  *DnsEntry  `yaml:"dnsEntry"`
}

func (s *Step) Ensure(ctx context.Context, command cmd.Factory) error {
	if s.Apply != "" {
		if err := command.Kubectl().ApplyFile(s.Apply).Cmd().Run(ctx); err != nil {
			return err
		}
	}
	if s.Delete != "" {
		if err := command.Kubectl().DeleteFile(s.Delete).Cmd().Run(ctx); err != nil {
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
		if err := s.DnsEntry.Ensure(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (s *Step) Teardown(ctx context.Context, command cmd.Factory) error {
	if s.Curl != nil {
		if err := s.Curl.Teardown(ctx, command); err != nil {
			return err
		}
	}
	if s.Condition != nil {
		if err := s.Condition.Teardown(ctx, command); err != nil {
			return err
		}
	}
	if s.DnsEntry != nil {
		if err := s.DnsEntry.Teardown(ctx, command); err != nil {
			return err
		}
	}
	return nil
}
