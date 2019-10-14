package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"go.uber.org/zap"
)

const (
	PetclinicDemoName                = "petclinic"
	PetclinicVirtualServiceName      = "default"
	PetclinicVirtualServiceNamespace = "gloo-system"
)

var (
	GlooctlNotProvidedError = errors.Errorf("must provide glooctl")
)

type Petclinic struct {
	DNS     *DNS `yaml:"dns"`
	Glooctl *Glooctl
}

func getFiles() []string {
	return []string{
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-vets.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-db.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-virtual-service.yaml",
	}
}

func (p *Petclinic) Ensure(ctx context.Context) error {
	if err := cmd.KubectlApplyAllFiles(ctx, getFiles()); err != nil {
		return err
	}
	if p.DNS != nil {
		if p.Glooctl == nil {
			return GlooctlNotProvidedError
		}
		proxyIp, err := p.Glooctl.GetProxyIp(ctx)
		if err != nil {
			return err
		}
		p.DNS.IP = proxyIp
		if p.DNS.Domain == "" {
			domain, err := internal.CreateDomain(ctx, PetclinicDemoName, p.DNS.HostedZone)
			if err != nil {
				return err
			}
			p.DNS.Domain = domain
		}
		err = p.DNS.Ensure(ctx)
		if err != nil {
			return err
		}
		return patchPetclinicVsWithDomain(ctx, p.DNS.Domain)
	}
	return nil
}

func (p *Petclinic) Teardown(ctx context.Context) error {
	if err := cmd.KubectlDeleteAllFiles(ctx, getFiles()); err != nil {
		return err
	}
	return nil
}

func patchPetclinicVsWithDomain(ctx context.Context, domain string) error {
	contextutils.LoggerFrom(ctx).Infow("Patching petclinic domain")
	patchStr := fmt.Sprintf("-p=[{\"op\":\"add\",\"path\":\"/spec/virtualHost/domains\",\"value\":[\"%s\"]}]", domain)
	out, err := cmd.
		Kubectl("patch", "vs", PetclinicVirtualServiceName).
		Namespace(PetclinicVirtualServiceNamespace).
		JsonPatch(patchStr).
		Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error patching petclinic virtualservice",
			zap.Error(err), zap.String("out", out), zap.String("domain", domain))
		return err
	}
	return nil
}
