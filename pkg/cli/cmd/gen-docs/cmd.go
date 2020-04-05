package gen_docs

import (
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/pkg/cli/options"
	"github.com/solo-io/valet/pkg/docs"
	"github.com/solo-io/valet/pkg/workflow"
	"github.com/spf13/cobra"
)

func GenDocs(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen-docs",
		Short: "generates docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return genDocs(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.PersistentFlags().StringVarP(&opts.GenDocs.Template, "template", "t", "", "path to markdown docs template")
	cmd.PersistentFlags().StringVarP(&opts.GenDocs.Output, "output", "o", "", "output path for markdown docs")

	return cmd
}

func genDocs(opts *options.Options) error {
	ctx := workflow.DefaultContext(opts.Top.Ctx)
	if opts.GenDocs.Template == "" {
		return errors.Errorf("Must provide input docs template")
	}
	if opts.GenDocs.Output == "" {
		return errors.Errorf("Must provide output file")
	}
	return docs.ProcessDoc(ctx, opts.GenDocs.Template, opts.GenDocs.Output)
}
