package gloo

import (
	"context"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
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
	return gloo.Install(ctx, config, localPathToGlooctl)
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