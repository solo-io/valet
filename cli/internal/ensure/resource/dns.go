package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/client"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type DNS struct {
	// The hosted zone to use for the DNS entry. This is required and must match the name of a hosted zone in Route53.
	// Valet will use the AWS credentials provided in AWS_SHARED_CREDENTIALS_FILE.
	// The credentials must have the following AWS privileges:
	//   route53:ChangeResourceRecordSets
	//   route53:ListHostedZones
	HostedZone string `yaml:"hostedZone"`
	Domain     string `yaml:"domain"`
	IP         string `yaml:"ip"`
}

func (d *DNS) Ensure(ctx context.Context, command cmd.Factory) error {
	awsClient, err := client.NewAwsDnsClient()
	if err != nil {
		return err
	}
	return awsClient.CreateMapping(ctx, d.HostedZone, d.Domain, d.IP)
}

func (d *DNS) Teardown(ctx context.Context, command cmd.Factory) error {
	return nil
}
