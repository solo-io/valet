package petclinic

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/cmd/ensure/gloo"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/options"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Petclinic(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "petclinic",
		Short:   "ensuring state of petclinic demo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return EnsurePetclinic(opts)
		},
	}

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func EnsurePetclinic(opts *options.Options) error {
	files := []string {
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-vets.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-db.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-virtual-service.yaml",
	}
	err := applyFiles(opts.Top.Ctx, files...)
	if err != nil {
		return err
	}

	if opts.Demos.Petclinic == nil || opts.Demos.Petclinic.DNS == nil {
		return nil
	}

	if opts.Demos.Petclinic.DNS.HostedZone == "" {
		contextutils.LoggerFrom(opts.Top.Ctx).Infow("No dns config provided")
	}

	client, err := gloo.NewAwsDnsClient()
	if err != nil {
		contextutils.LoggerFrom(opts.Top.Ctx).Errorw("Error creating aws dns client", zap.Error(err))
		return err
	}

	proxyIp, err := gloo.GetGlooProxyExternalIp(opts.Top.Ctx)
	if err != nil {
		return err
	}
	domain := opts.Demos.Petclinic.DNS.Domain
	if domain == "" {
		domain, err = internal.CreateDomain(opts.Top.Ctx, "petclinic", opts.Demos.Petclinic.DNS.HostedZone)
		if err != nil {
			return err
		}
	}
	err = client.CreateMapping(opts.Top.Ctx, opts.Demos.Petclinic.DNS.HostedZone, domain, proxyIp)
	if err != nil {
		return err
	}

	return patchPetclinicVsWithDomain(opts.Top.Ctx, domain)
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

func patchPetclinicVsWithDomain(ctx context.Context, domain string) error {
	contextutils.LoggerFrom(ctx).Infow("Patching petclinic domain")
	patchStr := fmt.Sprintf("-p=[{\"op\":\"add\",\"path\":\"/spec/virtualHost/domains\",\"value\":[\"%s\"]}]", domain)
	out, err := internal.ExecuteCmd("kubectl", "patch", "vs", "default", "-n", "gloo-system", "--type=json", patchStr)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error patching petclinic virtualservice",
			zap.Error(err), zap.String("out", out), zap.String("domain", domain))
		return err
	}
	return nil
}
