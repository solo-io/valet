package petclinic

import (
	"context"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func EnsurePetclinicDemoCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "petclinic",
		Short:   "ensuring state of petclinic demo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return EnsurePetclinicDemo(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func EnsurePetclinicDemo(opts *options.Options) error {
	files := []string {
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-vets.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-db.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-virtual-service.yaml",
	}
	return applyFiles(opts.Top.Ctx, files...)
}

func applyFiles(ctx context.Context, files ...string) error {
	for _, file := range files {
		err := applyFile(ctx, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyFile(ctx context.Context, file string) error {
	out, err := internal.ExecuteCmd("kubectl", "apply", "-f", file)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error applying file", zap.Error(err), zap.String("out", out), zap.String("file", file))
		return err
	}
	contextutils.LoggerFrom(ctx).Infow("Successfully applied file", zap.String("file", file))
	return nil
}