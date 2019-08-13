package gloo

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/solo-io/go-utils/errors"
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
	hostedZone, err := c.getHostedZone(ctx, "corp.solo.io.")
	if err != nil {
		return err
	}

	action := "CREATE"
	typeStr := "A"
	resourceRecord := &route53.ResourceRecord{
		Value: &ip,
	}
	resourceRecordSet := &route53.ResourceRecordSet{
		Type: &typeStr,
		Name: &domain,
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
	_, err = c.svc.ChangeResourceRecordSets(&input)
	return err
}