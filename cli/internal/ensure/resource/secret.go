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

func (s *Secret) Ensure(ctx context.Context) error {
	command := cmd.Kubectl().Create(secret).With(generic).WithName(s.Name).Namespace(s.Namespace)
	for name, v := range s.Entries {
		if v.File != "" {
			fromFile := fmt.Sprintf("--from-file=%s=%s", name, v.File)
			command = command.With(fromFile)
		} else if v.EnvVar != "" {
			fromLiteral := fmt.Sprintf("--from-literal=%s=%s", name, os.Getenv(v.EnvVar))
			command = command.With(fromLiteral)
		}
	}
	return command.DryRunAndApply(ctx)
}

func (s *Secret) Teardown(ctx context.Context) error {
	return cmd.Kubectl().Delete(secret).WithName(s.Name).Run(ctx)
}
