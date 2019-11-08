package ensure

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/teardown"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Application(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "application",
		Short: "ensures application is deployed",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Ensure.File == "" {
				return errors.Errorf("Must provide file to ensure")
			}
			err := ensureApplication(opts)
			if opts.Ensure.TeardownOnFinish {
				_ = teardown.TeardownApplication(opts)
			}
			return err
		},
	}
	cmd.PersistentFlags().StringToStringVarP(&opts.Ensure.Values, "values", "v", make(map[string]string), "values to provide to application")
	cmd.PersistentFlags().StringSliceVarP(&opts.Ensure.Flags, "flags", "", make([]string, 0), "flags to provide to application")
	cmd.PersistentFlags().BoolVarP(&opts.Ensure.DryRun, "dry-run", "", false, "dry-run and output rendered manifest to stdout")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func ensureApplication(opts *options.Options) error {
	if err := common.LoadEnv(opts.Top.Ctx); err != nil {
		return err
	}
	input := render.InputParams{
		Values: opts.Ensure.Values,
		Flags:  opts.Ensure.Flags,
		Step:   opts.Ensure.Step,
	}
	if opts.Ensure.File == "" {
		return common.MustProvideFileError
	}
	ref := application.Ref{
		Path: opts.Ensure.File,
	}
	command := cmd.CommandFactory{}
	if opts.Ensure.DryRun {
		manifest, err := renderManifest(opts.Top.Ctx, input, &command, ref)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", manifest)
		return nil
	}
	return ref.Ensure(opts.Top.Ctx, input, &command)
}

func renderManifest(ctx context.Context, input render.InputParams, command cmd.Factory, ref application.Ref) (string, error) {
	resources, err := ref.Render(ctx, input, command)
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	manifests, err := helmchart.ManifestsFromResources(resources)
	if err != nil {
		return "", err
	}
	return manifests.CombinedString() + "\n", nil
}
