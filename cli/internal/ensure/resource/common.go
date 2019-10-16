package resource

import (
	"context"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"os"
)

var (
	EnvVarNotFound = func(envVar string) error {
		return errors.Errorf("%s not found", envVar)
	}
)

const (
	AwsAccessKeyIdEnvVar        = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKeyEnvVar    = "AWS_SECRET_ACCESS_KEY"
	AwsAccessKeyIdSecretVar     = "access_key_id"
	AwsSecretAccessKeySecretVar = "secret_access_key"
)

func EnsureAll(ctx context.Context, command cmd.Factory, resources ...Resource) error {
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		if err := resource.Ensure(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func TeardownAll(ctx context.Context, command cmd.Factory, resources ...Resource) error {
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		if err := resource.Teardown(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func GetEnvVar(envVar string) (string, error) {
	val := os.Getenv(envVar)
	if val == "" {
		return "", EnvVarNotFound(envVar)
	}
	return val, nil
}

func AwsSecret(namespace, name string) *Secret {
	return &Secret{
		Name:      name,
		Namespace: namespace,
		Entries: map[string]SecretValue{
			AwsAccessKeyIdSecretVar:     {EnvVar: AwsAccessKeyIdEnvVar},
			AwsSecretAccessKeySecretVar: {EnvVar: AwsSecretAccessKeyEnvVar},
		},
	}
}
