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

var _ = Describe("curl", func() {
	const (
		host              = "test-host"
		path              = "/test-path"
		statusCode        = 123
		responseBody      = "body"
		serviceName       = "test-service"
		serviceNamespace  = "test-namespace"
		servicePort       = "test-port"
		ip                = "test-ip"
		otherStatusCode   = 321
		otherResponseBody = "other-body"
		attempts          = 5
		delay             = "10ms"
	)

	var (
		ctrl          *gomock.Controller
		runner        *mock_cmd.MockRunner
		ingressClient *mock_client.MockIngressClient
		input         render.InputParams

		ctx     = context.TODO()
		headers = map[string]string{
			"test": "header",
		}
		emptyErr = errors.Errorf("")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		ingressClient = mock_client.NewMockIngressClient(ctrl)
		input = render.InputParams{
			CommandRunner: runner,
			IngressClient: ingressClient,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("fully provided curl", func() {
		curl := &workflow.Curl{
			Host:         host,
			Path:         path,
			StatusCode:   statusCode,
			ResponseBody: responseBody,
			Headers:      headers,
			Service: workflow.ServiceRef{
				Name:      serviceName,
				Namespace: serviceNamespace,
				Port:      servicePort,
			},
			Attempts: attempts,
			Delay:    delay,
		}

		It("works for expected curl", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			req, err := curl.GetHttpRequest(ip)
			Expect(err).To(BeNil())
			runner.EXPECT().Request(ctx, gomock.Eq(req)).Return(responseBody, statusCode, nil)
			err = curl.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})

		It("returns error for unexpected response body", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			req, err := curl.GetHttpRequest(ip)
			Expect(err).To(BeNil())
			runner.EXPECT().Request(ctx, gomock.Eq(req)).Return(otherResponseBody, statusCode, nil).Times(attempts)
			err = curl.Ensure(ctx, input)
			Expect(err).To(Equal(workflow.UnexpectedResponseBodyError(otherResponseBody)))
		})

		It("returns error for unexpected status code", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			req, err := curl.GetHttpRequest(ip)
			Expect(err).To(BeNil())
			runner.EXPECT().Request(ctx, gomock.Eq(req)).Return(responseBody, otherStatusCode, nil).Times(attempts)
			err = curl.Ensure(ctx, input)
			Expect(err).To(Equal(workflow.UnexpectedStatusCodeError(otherStatusCode)))
		})

		It("returns error for request error", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			req, err := curl.GetHttpRequest(ip)
			Expect(err).To(BeNil())
			runner.EXPECT().Request(ctx, gomock.Eq(req)).Return(responseBody, statusCode, emptyErr).Times(attempts)
			err = curl.Ensure(ctx, input)
			Expect(err).To(Equal(emptyErr))
		})

		It("returns error for service error", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, emptyErr).Times(1)
			err := curl.Ensure(ctx, input)
			Expect(err).To(Equal(emptyErr))
		})
	})

	Context("curl default rendering", func() {
		curl := workflow.Curl{}

		It("works", func() {
			err := input.RenderFields(&curl)
			Expect(err).To(BeNil())
			err = input.RenderFields(&curl.Service)
			Expect(err).To(BeNil())
			Expect(curl.Attempts).To(Equal(workflow.DefaultCurlAttempts))
			Expect(curl.Delay).To(Equal(workflow.DefaultCurlDelay))
			Expect(curl.Service.Port).To(Equal(workflow.DefaultServicePort))
		})
	})

})
