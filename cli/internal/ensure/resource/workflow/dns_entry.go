package workflow

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

type DnsEntry struct {
	Domain string `yaml:"domain" valet:"key=Domain"`
	// This is "HostedZone" in AWS / Route53 DNS
	HostedZone string     `yaml:"hostedZone" valet:"key=HostedZone"`
	Service    ServiceRef `yaml:"service"`
}

func (d *DnsEntry) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	if err := input.Values.RenderFields(d); err != nil {
		return err
	}
	awsClient, err := client.NewAwsDnsClient()
	if err != nil {
		return err
	}
	ip, err := d.Service.getIp(ctx, input, command)
	if err != nil {
		return err
	}
	if err := awsClient.CreateMapping(ctx, d.HostedZone, d.Domain, ip); err != nil {
		return err
	}
	return nil
}

func (d *DnsEntry) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	cmd.Stderr().Println("Teardown not implemented")
	return nil
}

type ServiceRef struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Port      string `yaml:"port" valet:"default=http"`
}

func (s *ServiceRef) getAddress(ctx context.Context, input render.InputParams, command cmd.Factory) (string, error) {
	if err := input.Values.RenderFields(s); err != nil {
		return "", err
	}
	return client.GetIngressHost(s.Name, s.Namespace, s.Port)
}

func (s *ServiceRef) getIp(ctx context.Context, input render.InputParams, command cmd.Factory) (string, error) {
	url, err := s.getAddress(ctx, input, command)
	if err != nil {
		return "", err
	}
	parts := strings.Split(url, ":")
	if len(parts) <= 2 {
		return parts[0], nil
	}
	return "", errors.Errorf("Unexpected url %s", url)
}
