package cluster

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/kube-cluster/cli/options"
	"github.com/spf13/cobra"
)

var RootAddError = errors.Errorf("please select a subcommand")

func ClusterCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   "interacting with kube clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RootAddError
		},
	}

	cmd.AddCommand(EnsureCmd(opts))
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

