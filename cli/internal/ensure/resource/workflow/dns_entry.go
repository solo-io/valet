package workflow

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

const (
	DefaultServicePort = "http"
)

var (
	UnableToGetServiceIpError = func(err error) error {
		return errors.Wrapf(err, "unable to get service ip")
	}

	UnableToCreateDnsMappingError = func(err error) error {
		return errors.Wrapf(err, "unable to create dns mapping")
	}
)

type DnsEntry struct {
	Domain string `yaml:"domain" valet:"key=Domain"`
	// This is "HostedZone" in AWS / Route53 DNS
	HostedZone string     `yaml:"hostedZone" valet:"key=HostedZone"`
	Service    ServiceRef `yaml:"service"`
}

func (d *DnsEntry) Ensure(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(d); err != nil {
		return err
	}
	awsClient, err := input.GetDnsClient()
	if err != nil {
		return err
	}
	ip, err := d.Service.getIp(ctx, input)
	if err != nil {
		return UnableToGetServiceIpError(err)
	}
	if err := awsClient.CreateMapping(ctx, d.HostedZone, d.Domain, ip); err != nil {
		return UnableToCreateDnsMappingError(err)
	}
	return nil
}

func (d *DnsEntry) Teardown(ctx context.Context, input render.InputParams) error {
	cmd.Stderr().Println("Teardown not implemented")
	return nil
}

type ServiceRef struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace" valet:"key=Namespace"`
	Port      string `yaml:"port" valet:"default=http"`
}

func (s *ServiceRef) getAddress(ctx context.Context, input render.InputParams) (string, error) {
	if err := input.RenderFields(s); err != nil {
		return "", err
	}
	return input.GetIngressClient().GetIngressHost(s.Name, s.Namespace, s.Port)
}

func (s *ServiceRef) getIp(ctx context.Context, input render.InputParams) (string, error) {
	url, err := s.getAddress(ctx, input)
	if err != nil {
		return "", err
	}
	parts := strings.Split(url, ":")
	if len(parts) <= 2 {
		return parts[0], nil
	}
	return "", errors.Errorf("Unexpected url %s", url)
}
