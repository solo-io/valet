package aws

import (
	"fmt"
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/step/check"
	"strings"
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
	Domain string `json:"domain" valet:"key=Domain"`
	// This is "HostedZone" in AWS / Route53 DNS
	HostedZone string           `json:"hostedZone" valet:"key=HostedZone"`
	Service    check.ServiceRef `json:"service"`
}

func (d *DnsEntry) GetDescription(ctx *api.WorkflowContext, values render.Values) (string, error) {
	if err := values.RenderFields(d, ctx.Runner); err != nil {
		return "", err
	}
	ip, err := d.Service.GetIp(ctx, values)
	if err != nil {
		return "", UnableToGetServiceIpError(err)
	}
	return fmt.Sprintf("Creating DNS entry in AWS route53 for %s to %s", ip, strings.Join([]string{d.Domain, d.HostedZone}, ".")), nil
}

func (d *DnsEntry) Run(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(d, ctx.Runner); err != nil {
		return err
	}
	ip, err := d.Service.GetIp(ctx, values)
	if err != nil {
		return UnableToGetServiceIpError(err)
	}
	if err := ctx.AwsDnsClient.CreateMapping(ctx.Ctx, d.HostedZone, d.Domain, ip); err != nil {
		return UnableToCreateDnsMappingError(err)
	}
	return nil
}

func (d *DnsEntry) GetDocs(ctx *api.WorkflowContext, values render.Values, flags render.Flags) (string, error) {
	panic("implement me")
}
