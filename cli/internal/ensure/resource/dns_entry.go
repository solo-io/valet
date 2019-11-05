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

func (d *DnsEntry) updateWithValues(values Values) error {
	if d.Domain == "" {
		if values.ContainsKey(DomainKey) {
			if val, err := values.GetValue(DomainKey); err != nil {
				return err
			} else {
				d.Domain = val
			}
		}
	}
	if d.HostedZone == "" {
		if values.ContainsKey(HostedZoneKey) {
			if val, err := values.GetValue(HostedZoneKey); err != nil {
				return err
			} else {
				d.HostedZone = val
			}
		}
	}
	return nil
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
