package cluster

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/valet/cli/cmd/ensure/cluster/gke"
	"github.com/solo-io/valet/cli/cmd/ensure/cluster/minikube"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
)

func Cluster(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   "ensuring state of kube clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return internal.RootAddError
		},
	}

	cmd.AddCommand(
		gke.GkeCmd(opts),
		minikube.MinikubeCmd(opts))
	cmd.PersistentFlags().StringVarP(&opts.Cluster.KubeVersion, "kube-version", "v", "v1.13.0", "kube version")
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

