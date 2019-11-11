package teardown

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/config"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Application(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "application",
		Short: "tears down application based on configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return TeardownApplication(opts)
		},
	}

	cmd.PersistentFlags().StringToStringVarP(&opts.Ensure.Values, "values", "v", make(map[string]string), "values to provide to application")
	cmd.PersistentFlags().StringSliceVarP(&opts.Ensure.Flags, "flags", "", make([]string, 0), "flags to provide to application")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func TeardownApplication(opts *options.Options) error {
	globalConfig, err := config.LoadGlobalConfig(opts.Top.Ctx)
	if err != nil {
		return err
	}
	if err := common.LoadEnv(globalConfig); err != nil {
		return err
	}
	input := render.InputParams{
		Values:     opts.Ensure.Values,
		Flags:      opts.Ensure.Flags,
		Step:       opts.Ensure.Step,
		Registries: common.GetRegistries(globalConfig),
	}
	if opts.Ensure.File == "" {
		return common.MustProvideFileError
	}
	ref := application.Ref{
		Path: opts.Ensure.File,
	}
	command := cmd.CommandFactory{}
	return ref.Teardown(opts.Top.Ctx, input, &command)
}
