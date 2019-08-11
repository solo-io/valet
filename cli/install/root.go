package install

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/kube-cluster/cli/install/gloo"
	"github.com/solo-io/kube-cluster/cli/internal"
	"github.com/solo-io/kube-cluster/cli/options"
	"github.com/spf13/cobra"
)

func InstallCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "installing solo products on a kube cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return internal.RootAddError
		},
	}

	cmd.AddCommand(gloo.GlooCmd(opts))
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}
