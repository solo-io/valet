package gloo

import (
	"context"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

func GlooCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gloo",
		Short: "ensures gloo is installed to namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return EnsureGloo(opts)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.Gloo.Version, "version", "v", "", "gloo version")
	cmd.PersistentFlags().StringVarP(&opts.Gloo.Namespace, "namespace", "n", "gloo-system", "gloo namespace")
	cmd.PersistentFlags().BoolVarP(&opts.Gloo.Enterprise, "enterprise", "e", false, "install enterprise gloo")
	cmd.PersistentFlags().StringVar(&opts.Gloo.LicenseKey, "license-key", "", "enterprise gloo license key")

	cmd.PersistentFlags().BoolVarP(&opts.Gloo.AWS.Upstream, "upstream", "u", false, "create an AWS upstream from the AWS secret")
	cmd.PersistentFlags().BoolVarP(&opts.Gloo.AWS.Secret, "secret", "s", false, "create an AWS secret (requires ACCESS_KEY_ID and SECRET_ACCESS_KEY environment variables")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func EnsureGloo(opts *options.Options) error {
	if opts.Gloo.Namespace == "" {
		opts.Gloo.Namespace = "gloo-system"
	}
	if err := validateOpts(opts.Gloo); err != nil {
		return err
	}
	return ensureGloo(opts.Top.Ctx, opts.Gloo)
}

func ensureGloo(ctx context.Context, config options.Gloo) error {
	glooctl := NewGlooctlEnsurer()
	localPathToGlooctl, err := glooctl.Install(ctx, config)
	if err != nil {
		return err
	}

	gloo := NewGlooEnsurer()
	err = gloo.Install(ctx, config, localPathToGlooctl)
	if err != nil {
		return err
	}

	if !config.AWS.Secret {
		return nil
	}

	err = createAwsSecret(ctx, localPathToGlooctl)
	if err != nil {
		return err
	}

	if !config.AWS.Upstream {
		return nil
	}

	return createAwsUpstream(ctx, localPathToGlooctl)
}

func createAwsUpstream(ctx context.Context, localPathToGlooctl string) error {
	upstreamName := "aws"
	secretName := "aws-creds"
	_, err := internal.ExecuteCmd(localPathToGlooctl, "get", "upstream", upstreamName)
	if err == nil {
		// upstream already exists
		contextutils.LoggerFrom(ctx).Infow("aws upstream exists")
		return nil
	}
	out, err := internal.ExecuteCmd(localPathToGlooctl, "create", "upstream", upstreamName,
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

func createAwsSecret(ctx context.Context, localPathToGlooctl string) error {
	secretName := "aws-creds"
	_, err := internal.ExecuteCmd("kubectl", "get", "secret", "-n", "gloo-system", secretName)
	if err == nil {
		// secret already exists
		contextutils.LoggerFrom(ctx).Infow("aws secret exists")
		return nil
	}
	accessKey := os.Getenv("ACCESS_KEY_ID")
	if accessKey == "" {
		return errors.Errorf("Must specify ACCESS_KEY_ID in environment to create AWS secret")
	}
	secretKey := os.Getenv("SECRET_ACCESS_KEY")
	if secretKey == "" {
		return errors.Errorf("Must specify SECRET_ACCESS_KEY in environment to create AWS secret")
	}
	out, err := internal.ExecuteCmd(localPathToGlooctl, "create", "secret", "aws",
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

func validateOpts(config options.Gloo) error {
	if config.Version == "" {
		return errors.Errorf("must specify a version to install")
	}
	if config.Enterprise && config.LicenseKey == "" {
		return errors.Errorf("must specify a license-key when installing enterprise gloo")
	}
	return nil
}