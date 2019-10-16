package client

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"go.uber.org/zap"
)

var _ AwsDnsClient = new(awsDnsClient)

type AwsDnsClient interface {
	CreateMapping(ctx context.Context, hostedZoneName, domain, ip string) error
}

func NewAwsDnsClient() (*awsDnsClient, error) {
	config := aws.NewConfig()
	awsSession, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	svc := route53.New(awsSession)
	return &awsDnsClient{
		svc: svc,
	}, nil
}

type awsDnsClient struct {
	svc *route53.Route53
}

func (c *awsDnsClient) getHostedZone(ctx context.Context, name string) (*route53.HostedZone, error) {
	contextutils.LoggerFrom(ctx).Infow("Getting hosted zone id")
	listHostedZonesInput := route53.ListHostedZonesInput{}
	output, err := c.svc.ListHostedZones(&listHostedZonesInput)
	if err != nil {
		return nil, err
	}
	var hostedZone *route53.HostedZone
	for _, zone := range output.HostedZones {
		if *zone.Name == name {
			hostedZone = zone
		}
	}
	if hostedZone == nil {
		return nil, errors.Errorf("hosted zone not found")
	}
	return hostedZone, nil
}

func (c *awsDnsClient) CreateMapping(ctx context.Context, hostedZoneName, domain, ip string) error {
	hostedZone, err := c.getHostedZone(ctx, hostedZoneName)
	if err != nil {
		return err
	}

	action := "UPSERT"
	typeStr := "A"
	ttl := int64(300)
	resourceRecord := &route53.ResourceRecord{
		Value: &ip,
	}
	resourceRecordSet := &route53.ResourceRecordSet{
		Type: &typeStr,
		Name: &domain,
		TTL: &ttl,
		ResourceRecords: []*route53.ResourceRecord{resourceRecord},
	}
	change := &route53.Change{
		Action: &action,
		ResourceRecordSet: resourceRecordSet,
	}
	changeBatch := &route53.ChangeBatch{
		Changes: []*route53.Change{ change },
	}
	input := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: hostedZone.Id,
		ChangeBatch: changeBatch,
	}
	contextutils.LoggerFrom(ctx).Infow("Creating dns mapping",
		zap.String("hostedZone", hostedZoneName),
		zap.String("hostedZoneId", *hostedZone.Id),
		zap.String("domain", domain),
		zap.String("ip", ip))
	_, err = c.svc.ChangeResourceRecordSets(&input)
	return err
}