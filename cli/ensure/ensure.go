package ensure

import (
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/ensure/cluster"
	"github.com/solo-io/valet/cli/ensure/cluster/gke"
	"github.com/solo-io/valet/cli/ensure/cluster/minikube"
	"github.com/solo-io/valet/cli/ensure/demo"
	"github.com/solo-io/valet/cli/ensure/demo/petclinic"
	"github.com/solo-io/valet/cli/ensure/gloo"
	"github.com/solo-io/valet/cli/ensure/resources"
	"github.com/solo-io/valet/cli/file"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	MustProvideFileError = errors.Errorf("Must provide file option or subcommand")
)

func EnsureCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ensure",
		Short: "ensures kubernetes cluster is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ensure(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	cmd.PersistentFlags().StringVarP(&opts.Ensure.File, "file", "f", "", "path to file containing config to ensure")
	cmd.AddCommand(
		cluster.ClusterCmd(opts, optionsFunc...),
		gloo.GlooCmd(opts, optionsFunc...),
		demo.DemoCmd(opts, optionsFunc...),
		resources.ResourcesCmd(opts, optionsFunc...))
	return cmd
}

func ensure(opts *options.Options) error {
	if opts.Ensure.File == "" {
		return MustProvideFileError
	}

	config, err := file.LoadConfig(opts.Top.Ctx, opts.Ensure.File)
	if err != nil {
		return err
	}

	if config.Cluster != nil {
		opts.Cluster.Type = config.Cluster.Type
		opts.Cluster.GKE = config.Cluster.GKE
		opts.Cluster.Minikube = config.Cluster.Minikube

		var clusterErr error
		if opts.Cluster.Type == "gke" {
			clusterErr = gke.EnsureGke(opts)
		} else if opts.Cluster.Type == "minikube" {
			clusterErr = minikube.EnsureMinikube(opts)
		} else {
			return errors.Errorf("unknown type", zap.String("type", opts.Cluster.Type))
		}
		if clusterErr != nil {
			return clusterErr
		}
	}

	if config.Gloo != nil {
		opts.Gloo = *config.Gloo
		err := gloo.EnsureGloo(opts)
		if err != nil {
			return err
		}
	}

	if config.Demos != nil {
		if config.Demos.Petclinic != nil {
			opts.Demos.Petclinic = config.Demos.Petclinic
			err := petclinic.EnsurePetclinicDemo(opts)
			if err != nil {
				return err
			}
		}
	}

	if config.Resources != nil {
		opts.Ensure.Resources = config.Resources
		err := resources.EnsureResources(opts)
		if err != nil {
			return err
		}
	}

	return nil
}
