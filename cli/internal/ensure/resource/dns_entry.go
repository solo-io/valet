package resource

import (
	"context"

	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type DnsEntry struct {
	Domain string `yaml:"domain"`
	// This is "HostedZone" in AWS / Route53 DNS
	HostedZone string     `yaml:"hostedZone"`
	Service    ServiceRef `yaml:"service"`
}

func (d *DnsEntry) updateWithValues(values map[string]string) {
	if d.Domain == "" {
		if val, ok := values[DomainKey]; ok {
			d.Domain = val
		}
	}
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
	if err := awsClient.CreateMapping(ctx, d.HostedZone, d.Domain, ip); err != nil {
		return err
	}
	return nil
}

func (d *DnsEntry) Teardown(ctx context.Context, command cmd.Factory) error {
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
