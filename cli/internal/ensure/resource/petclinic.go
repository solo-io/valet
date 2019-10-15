package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	PetclinicDemoName                = "petclinic"
	PetclinicVirtualServiceName      = "default"
	PetclinicVirtualServiceNamespace = "gloo-system"
)

var (
	GlooctlNotProvidedError = errors.Errorf("must provide glooctl")
	petclinicFiles          = []string{
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-vets.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-db.yaml",
		"https://raw.githubusercontent.com/sololabs/demos/master/petclinic_demo/petclinic-virtual-service.yaml",
	}
	petclinicResources = Resources{Paths: petclinicFiles}
)

type Petclinic struct {
	DNS *DNS `yaml:"dns"`
}

func (p *Petclinic) Ensure(ctx context.Context, command cmd.Factory) error {
	if err := petclinicResources.Ensure(ctx, command); err != nil {
		return err
	}
	if p.DNS != nil {
		proxyIp, err := command.Glooctl().GetProxyIp(ctx)
		if err != nil {
			return err
		}
		p.DNS.IP = proxyIp
		if p.DNS.Domain == "" {
			domain, err := internal.CreateDomainString(ctx, PetclinicDemoName, p.DNS.HostedZone)
			if err != nil {
				return err
			}
			p.DNS.Domain = domain
		}
		err = p.DNS.Ensure(ctx, command)
		if err != nil {
			return err
		}
		return patchPetclinicVsWithDomain(ctx, command, p.DNS.Domain)
	}
	return nil
}

func (p *Petclinic) Teardown(ctx context.Context, command cmd.Factory) error {
	return petclinicResources.Teardown(ctx, command)
}

func patchPetclinicVsWithDomain(ctx context.Context, command cmd.Factory, domain string) error {
	contextutils.LoggerFrom(ctx).Infow("patching petclinic domain")
	patchStr := fmt.Sprintf("-p=[{\"op\":\"add\",\"path\":\"/spec/virtualHost/domains\",\"value\":[\"%s\"]}]", domain)
	return command.
		Kubectl().
		With("patch", "vs", PetclinicVirtualServiceName).
		Namespace(PetclinicVirtualServiceNamespace).
		JsonPatch(patchStr).
		Run(ctx)
}
