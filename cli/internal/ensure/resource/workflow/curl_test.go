package workflow_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	errors "github.com/rotisserie/eris"
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
			Service: &workflow.ServiceRef{
				Name:      serviceName,
				Namespace: serviceNamespace,
				Port:      servicePort,
			},
			Attempts: attempts,
			Delay:    delay,
		}

		fullUrl := fmt.Sprintf("%s://%s%s", servicePort, ip, path)

		It("works for expected curl", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			req, err := curl.GetHttpRequest(fullUrl)
			Expect(err).To(BeNil())
			runner.EXPECT().Request(ctx, gomock.Eq(req)).Return(responseBody, statusCode, nil)
			err = curl.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})

		It("returns error for unexpected response body", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			req, err := curl.GetHttpRequest(fullUrl)
			Expect(err).To(BeNil())
			runner.EXPECT().Request(ctx, gomock.Eq(req)).Return(otherResponseBody, statusCode, nil).Times(attempts)
			err = curl.Ensure(ctx, input)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal(workflow.UnexpectedResponseBodyError(otherResponseBody).Error()))
		})

		It("returns error for unexpected status code", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			req, err := curl.GetHttpRequest(fullUrl)
			Expect(err).To(BeNil())
			runner.EXPECT().Request(ctx, gomock.Eq(req)).Return(responseBody, otherStatusCode, nil).Times(attempts)
			err = curl.Ensure(ctx, input)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal(workflow.UnexpectedStatusCodeError(otherStatusCode).Error()))
		})

		It("returns error for request error", func() {
			ingressClient.EXPECT().GetIngressHost(serviceName, serviceNamespace, servicePort).Return(ip, nil).Times(1)
			req, err := curl.GetHttpRequest(fullUrl)
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
		curl := workflow.Curl{
			Service: &workflow.ServiceRef{},
			PortForward: &workflow.PortForward{},
		}

		It("works", func() {
			err := input.RenderFields(&curl)
			Expect(err).To(BeNil())
			err = input.RenderFields(curl.Service)
			Expect(err).To(BeNil())
			err = input.RenderFields(curl.PortForward)
			Expect(err).To(BeNil())
			Expect(curl.Attempts).To(Equal(workflow.DefaultCurlAttempts))
			Expect(curl.Delay).To(Equal(workflow.DefaultCurlDelay))
			Expect(curl.Method).To(Equal(workflow.DefaultMethod))
			Expect(curl.Service.Port).To(Equal(workflow.DefaultServicePort))
			Expect(curl.PortForward.Port).To(Equal(workflow.DefaultPortForwardPort))
		})
	})

})
