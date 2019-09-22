package gloo

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/internal"
	"go.uber.org/zap"
	"os"
)

func EnsureGloo(ctx context.Context, glooManager GlooManager, gloo *api.Gloo) error {
	if err := validateConfig(ctx, gloo); err != nil {
		return err
	}
	if err := glooManager.Install(ctx); err != nil {
		return err
	}
	if err := createAwsResources(ctx, gloo, glooManager); err != nil {
		return err
	}

	if gloo.UiVirtualService != nil {
		vsCreator := NewKubectlUiVirtualServiceCreator()
		return vsCreator.Create(ctx, *gloo.UiVirtualService)
	}
	return nil
}

func createAwsResources(ctx context.Context, config *api.Gloo, gloo GlooManager) error {
	if !config.AWS.Secret {
		return nil
	}

	err := createAwsSecret(ctx, gloo)
	if err != nil {
		return err
	}

	if !config.AWS.Upstream {
		return nil
	}

	err = createAwsUpstream(ctx, gloo)
	if err != nil {
		return err
	}
	return nil
}

func createAwsUpstream(ctx context.Context, gloo GlooManager) error {
	upstreamName := "aws"
	secretName := "aws-creds"
	_, err := gloo.Glooctl().Execute("get", "upstream", upstreamName)
	if err == nil {
		// upstream already exists
		contextutils.LoggerFrom(ctx).Infow("aws upstream exists")
		return nil
	}
	out, err := gloo.Glooctl().Execute("create", "upstream", upstreamName,
		"--aws-secret-name", secretName,
		"--name", upstreamName)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error trying to create aws upstream",
			zap.Error(err),
			zap.String("out", out))
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("created aws upstream")
	return nil
}

func createAwsSecret(ctx context.Context, gloo GlooManager) error {
	secretName := "aws-creds"
	_, err := internal.ExecuteCmd("kubectl", "get", "secret", "-n", "gloo-system", secretName)
	if err == nil {
		// secret already exists
		contextutils.LoggerFrom(ctx).Infow("aws secret exists")
		return nil
	}
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKey == "" {
		return errors.Errorf("Must specify AWS_ACCESS_KEY_ID in environment to create AWS secret")
	}
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretKey == "" {
		return errors.Errorf("Must specify AWS_SECRET_ACCESS_KEY in environment to create AWS secret")
	}
	out, err := gloo.Glooctl().Execute("create", "secret", "aws",
		"--secret-key", secretKey,
		"--access-key", accessKey,
		"--name", secretName)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error trying to create aws secret",
			zap.Error(err),
			zap.String("out", out))
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("created aws secret")
	return nil
}

func validateConfig(ctx context.Context, config *api.Gloo) error {
	if config.Enterprise && config.LicenseKey == "" {
		if os.Getenv("LICENSE_KEY") != "" {
			config.LicenseKey = os.Getenv("LICENSE_KEY")
		} else {
			return errors.Errorf("must specify a license-key when installing enterprise gloo")
		}
	}
	return nil
}
