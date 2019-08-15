package gloo

import (
	"context"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
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
	cmd.PersistentFlags().BoolVarP(&opts.Gloo.Enterprise, "enterprise", "e", false, "install enterprise gloo")
	cmd.PersistentFlags().StringVar(&opts.Gloo.LicenseKey, "license-key", "", "enterprise gloo license key")

	cmd.PersistentFlags().BoolVarP(&opts.Gloo.AWS.Upstream, "upstream", "u", false, "create an AWS upstream from the AWS secret")
	cmd.PersistentFlags().BoolVarP(&opts.Gloo.AWS.Secret, "secret", "s", false, "create an AWS secret (requires AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables")

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func EnsureGloo(opts *options.Options) error {
	if err := validateOpts(opts.Top.Ctx, &opts.Gloo); err != nil {
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

	err = createAwsResources(ctx, config, localPathToGlooctl)
	if err != nil {
		return err
	}

	if config.UiVirtualService != nil {
		vsCreator := NewKubectlUiVirtualServiceCreator()
		return vsCreator.Create(ctx, *config.UiVirtualService)
	}
	return nil
}

func createAwsResources(ctx context.Context, config options.Gloo, localPathToGlooctl string) error {
	if !config.AWS.Secret {
		return nil
	}

	err := createAwsSecret(ctx, localPathToGlooctl)
	if err != nil {
		return err
	}

	if !config.AWS.Upstream {
		return nil
	}

	err = createAwsUpstream(ctx, localPathToGlooctl)
	if err != nil {
		return err
	}
	return nil
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
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKey == "" {
		return errors.Errorf("Must specify AWS_ACCESS_KEY_ID in environment to create AWS secret")
	}
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretKey == "" {
		return errors.Errorf("Must specify AWS_SECRET_ACCESS_KEY in environment to create AWS secret")
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

func validateOpts(ctx context.Context, config *options.Gloo) error {
	if config.Version == "" {
		tag, err := getLatestTag(ctx, config)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("Error determining latest release", zap.Error(err))
			return errors.Errorf("did not specify a version to install, and couldn't determine latest release")
		}
		contextutils.LoggerFrom(ctx).Infow("Setting version to latest release", zap.String("tag", tag))
		config.Version = tag
	}
	if config.Enterprise && config.LicenseKey == "" {
		if os.Getenv("LICENSE_KEY") != "" {
			config.LicenseKey = os.Getenv("LICENSE_KEY")
		} else {
			return errors.Errorf("must specify a license-key when installing enterprise gloo")
		}
	}
	return nil
}

func getLatestTag(ctx context.Context, config *options.Gloo) (string, error) {
	client := githubutils.GetClientWithOrWithoutToken(ctx)
	repo := getRepo(config.Enterprise)
	release, _, err := client.Repositories.GetLatestRelease(ctx, "solo-io", repo)
	if err != nil {
		return "", err
	}
	return release.GetTagName(), nil
}
