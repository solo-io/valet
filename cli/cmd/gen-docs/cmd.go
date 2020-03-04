package gen_docs

import (
	"fmt"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/internal/ensure/resource/workflow"
	"github.com/solo-io/valet/cli/options"
	"github.com/solo-io/valet/cli/version"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

func GenDocs(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen-docs",
		Short: "generates docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setContext(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.PersistentFlags().StringVarP(&opts.Ensure.File, "file", "f", "", "path to file containing config to ensure")
	cmd.PersistentFlags().StringVarP(&opts.GenDocs.Output, "output", "o", "", "path to output file")

	return cmd
}

func setContext(opts *options.Options) error {
	input, err := common.LoadInput(opts)
	if err != nil {
		return err
	}
	cfg, err := common.LoadConfig(opts, *input)
	if err != nil {
		return err
	}
	docs := workflow.Section{}
	err = cfg.Document(opts.Top.Ctx, *input, &docs)
	if err != nil {
		return err
	}
	markdown := toMarkdown(&docs, 1)
	markdown = getHeader(opts) + "\n\n" + markdown
	if opts.GenDocs.Output != "" {
		ioutil.WriteFile(opts.GenDocs.Output, []byte(markdown), os.ModePerm)
	} else {
		fmt.Printf("%v", markdown)
	}
	return nil
}

func getHeader(opts *options.Options) string {
	return fmt.Sprintf("_This doc was automatically created by Valet %s from the workflow defined in %s. To deploy the demo, you can use `valet ensure -f %s` from this directory, " +
		"or execute the steps manually. Do not modify this file directly, it will be overwritten the next time the docs are generated._",
		version.Version, opts.Ensure.File, opts.Ensure.File)
}

func toMarkdown(section *workflow.Section, indent int) string {
	title := " " + section.Title
	if section.Title != "" {
		for i := 1; i <= indent; i++ {
			title = "#" + title
		}
	}

	body := title + "\n\n" +  section.Description
	if section.Notes != "" {
		body = body + "\n\n" + section.Notes
	}

	for _, subsection := range section.Sections {
		body = body + "\n\n" + toMarkdown(&subsection, indent+1)
	}
	return body
}
