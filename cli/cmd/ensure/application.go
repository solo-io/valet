package ensure

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/cliutils"
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/teardown"
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
	input, err := common.LoadInput(opts)
	if err != nil {
		return err
	}
	ref := application.Ref{
		RegistryName: opts.Ensure.Registry,
		Path:         opts.Ensure.File,
	}
	if opts.Ensure.DryRun {
		return renderManifest(opts.Top.Ctx, *input, ref)
	}
	return ref.Ensure(opts.Top.Ctx, *input)
}

func renderManifest(ctx context.Context, input render.InputParams, ref application.Ref) error {
	resources, err := ref.Render(ctx, input)
	if err != nil {
		return err
	}
	manifests, err := helmchart.ManifestsFromResources(resources)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", manifests.CombinedString()+"\n")
	return nil
}
