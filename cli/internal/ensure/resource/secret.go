package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"os"
)

const (
	secret  = "secret"
	generic = "generic"
)

var _ Resource = new(Secret)

type Secret struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Entries   map[string]SecretValue
}

type SecretValue struct {
	EnvVar string `yaml:"envVar"`
	File   string `yaml:"file"`
}

func (s *Secret) Ensure(ctx context.Context, command cmd.Factory) error {
	toRun := command.Kubectl().Create(secret).With(generic).WithName(s.Name).Namespace(s.Namespace)
	for name, v := range s.Entries {
		if v.File != "" {
			fromFile := fmt.Sprintf("--from-file=%s=%s", name, v.File)
			toRun = toRun.With(fromFile)
		} else if v.EnvVar != "" {
			template := "--from-literal=%s=%s"
			fromLiteral := fmt.Sprintf(template, name, os.Getenv(v.EnvVar))
			fromLiteralRedacted := fmt.Sprintf(template, name, cmd.Redacted)
			toRun = toRun.With(fromLiteral).Redact(fromLiteral, fromLiteralRedacted)
		}
	}
	return toRun.DryRunAndApply(ctx, command)
}

func (s *Secret) Teardown(ctx context.Context, command cmd.Factory) error {
	return command.Kubectl().Delete(secret).WithName(s.Name).Cmd().Run(ctx)
}
