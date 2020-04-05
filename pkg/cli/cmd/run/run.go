package run

import (
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/pkg/cli/options"
	"github.com/solo-io/valet/pkg/workflow"
	"github.com/spf13/cobra"
)

func Run(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "run a valet workflow",
		RunE: func(_ *cobra.Command, args []string) error {
			return run(opts)
		},
	}

	cliutils.ApplyOptions(runCmd, optionsFunc)
	runCmd.PersistentFlags().StringVarP(&opts.Run.File, "file", "f", "", "path to file containing config to ensure")
	runCmd.PersistentFlags().StringToStringVarP(&opts.Run.Values, "values", "v", make(map[string]string), "values to provide to workflow")
	return runCmd
}

func run(opts *options.Options) error {
	if opts.Run.File == "" {
		return errors.Errorf("Must provide file containing yaml workflow")
	}
	ctx := workflow.DefaultContext(opts.Top.Ctx)
	toRun := workflow.Workflow{}
	if err := ctx.FileStore.LoadYaml(opts.Run.File, &toRun); err != nil {
		return err
	}
	toRun.Values = toRun.Values.MergeValues(opts.Run.Values)
	return toRun.Run(ctx)
}
