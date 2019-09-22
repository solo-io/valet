package petclinic

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/gloo"
	"go.uber.org/zap"
)

func EnsurePetclinic(ctx context.Context, petclinic *api.Petclinic) error {
	files := []string{
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-vets.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-db.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-virtual-service.yaml",
	}
	err := applyFiles(ctx, files...)
	if err != nil {
		return err
	}

	if petclinic == nil || petclinic.DNS == nil {
		return nil
	}

	if petclinic.DNS.HostedZone == "" {
		contextutils.LoggerFrom(ctx).Infow("No dns config provided")
	}

	client, err := gloo.NewAwsDnsClient()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error creating aws dns client", zap.Error(err))
		return err
	}

	proxyIp, err := gloo.GetGlooProxyExternalIp(ctx)
	if err != nil {
		return err
	}
	domain := petclinic.DNS.Domain
	if domain == "" {
		domain, err = internal.CreateDomain(ctx, "petclinic", petclinic.DNS.HostedZone)
		if err != nil {
			return err
		}
	}
	err = client.CreateMapping(ctx, petclinic.DNS.HostedZone, domain, proxyIp)
	if err != nil {
		return err
	}

	return patchPetclinicVsWithDomain(ctx, domain)
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
