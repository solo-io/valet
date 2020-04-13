package aws_test

import (
	"github.com/solo-io/valet/pkg/api"
	mock_aws "github.com/solo-io/valet/pkg/client/aws/mocks"
	mock_kube "github.com/solo-io/valet/pkg/client/kube/mocks"
	mock_cmd "github.com/solo-io/valet/pkg/cmd/mocks"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/step/aws"
	"github.com/solo-io/valet/pkg/step/check"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	errors "github.com/rotisserie/eris"
)

var _ = Describe("dns", func() {
	const (
		domain     = "test-domain"
		hostedZone = "test-hosted-zone"

		serviceName      = "test-service"
		serviceNamespace = "test-namespace"
		servicePort      = "test-port"
		ip               = "test-ip"
		address          = "test-ip:80"
	)

	var (
		ctrl         *gomock.Controller
		runner       *mock_cmd.MockRunner
		kubeClient   *mock_kube.MockClient
		awsDnsClient *mock_aws.MockDnsClient
		ctx          *api.WorkflowContext

		emptyErr = errors.Errorf("")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		awsDnsClient = mock_aws.NewMockDnsClient(ctrl)
		kubeClient = mock_kube.NewMockClient(ctrl)
		ctx = &api.WorkflowContext{
			Runner:       runner,
			KubeClient:   kubeClient,
			AwsDnsClient: awsDnsClient,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("fully provided dns entry", func() {
		dns := aws.DnsEntry{
			Domain: domain,
			Service: check.ServiceRef{
				Port:      servicePort,
				Namespace: serviceNamespace,
				Name:      serviceName,
			},
			HostedZone: hostedZone,
		}

		It("works for expected dns entry", func() {
			kubeClient.EXPECT().GetIngressAddress(serviceName, serviceNamespace, servicePort).Return(address, nil).Times(1)
			awsDnsClient.EXPECT().CreateMapping(nil, hostedZone, domain, ip).Return(nil).Times(1)
			err := dns.Run(ctx, nil)
			Expect(err).To(BeNil())
		})

		It("returns error if service ip can't be determined", func() {
			kubeClient.EXPECT().GetIngressAddress(serviceName, serviceNamespace, servicePort).Return("", emptyErr).Times(1)
			err := dns.Run(ctx, nil)
			Expect(err).To(Equal(aws.UnableToGetServiceIpError(emptyErr)))
		})

		It("returns error if creating mapping fails", func() {
			kubeClient.EXPECT().GetIngressAddress(serviceName, serviceNamespace, servicePort).Return(address, nil).Times(1)
			awsDnsClient.EXPECT().CreateMapping(nil, hostedZone, domain, ip).Return(emptyErr).Times(1)
			err := dns.Run(ctx, nil)
			Expect(err).To(Equal(aws.UnableToCreateDnsMappingError(emptyErr)))
		})

	})

	Context("dns values rendering", func() {
		dns := aws.DnsEntry{
			Service: check.ServiceRef{
				Namespace: serviceNamespace,
				Name:      serviceName,
			},
		}

		values := render.Values{
			render.HostedZoneKey: hostedZone,
			render.DomainKey:     domain,
		}

		It("works", func() {
			err := values.RenderFields(&dns, ctx.Runner)
			Expect(err).To(BeNil())
			err = values.RenderFields(&dns.Service, ctx.Runner)
			Expect(err).To(BeNil())
			Expect(dns.HostedZone).To(Equal(hostedZone))
			Expect(dns.Domain).To(Equal(domain))
			Expect(dns.HostedZone).To(Equal(hostedZone))
			Expect(dns.Service.Port).To(Equal(aws.DefaultServicePort))
		})
	})

})
