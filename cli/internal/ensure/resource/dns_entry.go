package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type DnsEntry struct {
	Subdomain string `yaml:"subdomain"`
	// This is "HostedZone" in AWS / Route53 DNS
	HostedZone string     `yaml:"hostedZone"`
	Service    ServiceRef `yaml:"service"`
}

func (d *DnsEntry) Ensure(ctx context.Context, command cmd.Factory) error {
	awsClient, err := client.NewAwsDnsClient()
	if err != nil {
		return err
	}
	ip, err := d.Service.getIpAddress(ctx, command)
	if err != nil {
		return err
	}
	fullDomain := fmt.Sprintf("%s.%s", d.Subdomain, d.HostedZone)
	if err := awsClient.CreateMapping(ctx, d.HostedZone, fullDomain, ip); err != nil {
		return err
	}
	return nil
}

func (d *DnsEntry) Teardown(ctx context.Context, command cmd.Factory) error {
	contextutils.LoggerFrom(ctx).Errorw("Teardown not implemented")
	return nil
}

type ServiceRef struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

func (s *ServiceRef) getIpAddress(ctx context.Context, command cmd.Factory) (string, error) {
	return command.Kubectl().GetServiceIP(ctx, s.Namespace, s.Name)
}
