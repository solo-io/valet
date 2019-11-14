package workflow_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/errors"
	mock_client "github.com/solo-io/valet/cli/internal/ensure/client/mocks"
	mock_cmd "github.com/solo-io/valet/cli/internal/ensure/cmd/mocks"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"github.com/solo-io/valet/cli/internal/ensure/resource/workflow"
)

var _ = Describe("dns", func() {
	const (
		domain     = "test-domain"
		hostedZone = "test-hosted-zone"

		serviceName      = "test-service"
		serviceNamespace = "test-namespace"
		servicePort      = "test-port"
		ip               = "test-ip"
	)

	var (
		ctrl          *gomock.Controller
		runner        *mock_cmd.MockRunner
		ingressClient *mock_client.MockIngressClient
		awsDnsClient  *mock_client.MockAwsDnsClient
		input         render.InputParams

		ctx      = context.TODO()
		emptyErr = errors.Errorf("")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		awsDnsClient = mock_client.NewMockAwsDnsClient(ctrl)
		ingressClient = mock_client.NewMockIngressClient(ctrl)
		input = render.InputParams{
			CommandRunner: runner,
			DnsClient:     awsDnsClient,
			IngressClient: ingressClient,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("fully provided dns entry", func() {
		dns := workflow.DnsEntry{
			Domain: domain,
			Service: workflow.ServiceRef{
				Port:      servicePort,
				Namespace: serviceNamespace,
				Name:      serviceName,
			},
			HostedZone: hostedZone,
		}

		It("works for expected dns entry", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			awsDnsClient.EXPECT().CreateMapping(ctx, hostedZone, domain, ip).Return(nil).Times(1)
			err := dns.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})

		It("returns error if service ip can't be determined", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return("", emptyErr).Times(1)
			err := dns.Ensure(ctx, input)
			Expect(err).To(Equal(workflow.UnableToGetServiceIpError(emptyErr)))
		})

		It("returns error if creating mapping fails", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			awsDnsClient.EXPECT().CreateMapping(ctx, hostedZone, domain, ip).Return(emptyErr).Times(1)
			err := dns.Ensure(ctx, input)
			Expect(err).To(Equal(workflow.UnableToCreateDnsMappingError(emptyErr)))
		})

	})

	Context("dns values rendering", func() {
		dns := workflow.DnsEntry{
			Service: workflow.ServiceRef{
				Namespace: serviceNamespace,
				Name:      serviceName,
			},
		}

		BeforeEach(func() {
			values := render.Values{
				render.HostedZoneKey: hostedZone,
				render.DomainKey:     domain,
			}
			input.Values = values
		})

		It("works", func() {
			err := input.RenderFields(&dns)
			Expect(err).To(BeNil())
			err = input.RenderFields(&dns.Service)
			Expect(err).To(BeNil())
			Expect(dns.HostedZone).To(Equal(hostedZone))
			Expect(dns.Domain).To(Equal(domain))
			Expect(dns.HostedZone).To(Equal(hostedZone))
			Expect(dns.Service.Port).To(Equal(workflow.DefaultServicePort))
		})
	})

})
