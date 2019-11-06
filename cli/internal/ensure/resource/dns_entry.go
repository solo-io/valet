package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type DnsEntry struct {
	Domain string `yaml:"domain" valet:"key=Domain"`
	// This is "HostedZone" in AWS / Route53 DNS
	HostedZone string     `yaml:"hostedZone" valet:"key=HostedZone"`
	Service    ServiceRef `yaml:"service"`
}

func (d *DnsEntry) Ensure(ctx context.Context, input InputParams, command cmd.Factory) error {
	if err := input.Values.RenderFields(d); err != nil {
		return err
	}
	awsClient, err := client.NewAwsDnsClient()
	if err != nil {
		return err
	}
	ip, err := d.Service.getIpAddress(ctx, command)
	if err != nil {
		return err
	}
	if err := awsClient.CreateMapping(ctx, d.HostedZone, d.Domain, ip); err != nil {
		return err
	}
	return nil
}

func (d *DnsEntry) Teardown(ctx context.Context, input InputParams, command cmd.Factory) error {
	cmd.Stderr().Println("Teardown not implemented")
	return nil
}

type ServiceRef struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

func (s *ServiceRef) getIpAddress(ctx context.Context, command cmd.Factory) (string, error) {
	return command.Kubectl().GetServiceIP(ctx, s.Namespace, s.Name)
}
