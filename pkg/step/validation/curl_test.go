package validation_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/pkg/api"
	mock_kube "github.com/solo-io/valet/pkg/client/kube/mocks"
	"github.com/solo-io/valet/pkg/cmd"
	mock_cmd "github.com/solo-io/valet/pkg/cmd/mocks"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/step/validation"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var _ = Describe("apply", func() {

	const (
		path    = "/path"
		host    = "host"
		svcName = "gateway-proxy"
		svcNs   = "gloo-system"
		svcPort = "http"
	)

	var (
		ctrl            *gomock.Controller
		runner          *mock_cmd.MockRunner
		kubeClient      *mock_kube.MockClient
		ctx             *api.WorkflowContext
		gatewayProxySvc = &validation.ServiceRef{Namespace: svcNs, Name: svcName}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		kubeClient = mock_kube.NewMockClient(ctrl)
		ctx = &api.WorkflowContext{
			Runner:     runner,
			KubeClient: kubeClient,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("runs with default method", func() {
		curl := validation.Curl{
			Path:       path,
			Host:       host,
			Service:    gatewayProxySvc,
			StatusCode: 200,
		}
		kubeClient.EXPECT().GetIngressHost(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := http.NewRequest(validation.DefaultMethod, "http://host/path", nil)
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs with the right default attempts during status code failure", func() {
		curl := validation.Curl{
			Path:       path,
			Host:       host,
			Service:    gatewayProxySvc,
			StatusCode: 200,
			Delay:      "1ms",
		}
		kubeClient.EXPECT().GetIngressHost(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := http.NewRequest(validation.DefaultMethod, "http://host/path", nil)
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("", 503, nil).Times(10)
		err = curl.Run(ctx, nil)
		Expect(err.Error()).To(Equal(validation.UnexpectedStatusCodeError(503).Error()))
	})

	It("runs with a specific set of attempts during body failure", func() {
		curl := validation.Curl{
			Path:         path,
			Host:         host,
			Service:      gatewayProxySvc,
			StatusCode:   200,
			Delay:        "1ms",
			Attempts:     1,
			ResponseBody: "foo",
		}
		kubeClient.EXPECT().GetIngressHost(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := http.NewRequest(validation.DefaultMethod, "http://host/path", nil)
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("bar", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err.Error()).To(Equal(validation.UnexpectedResponseBodyError("bar").Error()))
	})

	It("runs checking response body", func() {
		curl := validation.Curl{
			Path:         path,
			Host:         host,
			Service:      gatewayProxySvc,
			StatusCode:   200,
			ResponseBody: "bar",
		}
		kubeClient.EXPECT().GetIngressHost(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := curl.GetHttpRequest("http://host/path")
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("bar", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs checking response body substring", func() {
		curl := validation.Curl{
			Path:                  path,
			Host:                  host,
			Service:               gatewayProxySvc,
			StatusCode:            200,
			ResponseBodySubstring: "bar",
		}
		kubeClient.EXPECT().GetIngressHost(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := curl.GetHttpRequest("http://host/path")
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("barfoo", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs with default status code 200", func() {
		curl := validation.Curl{
			Path:                  path,
			Host:                  host,
			Service:               gatewayProxySvc,
		}
		kubeClient.EXPECT().GetIngressHost(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := curl.GetHttpRequest("http://host/path")
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs with a port forward fully specified", func() {
		portFwd := validation.PortForward{
			Namespace:      "ns",
			DeploymentName: "dep",
			Port:           1234,
		}
		curl := validation.Curl{
			Path:                  path,
			Host:                  "localhost",
			PortForward:           &portFwd,
			StatusCode:            200,
		}
		req, err := curl.GetHttpRequest("http://localhost:1234/path")
		Expect(err).To(BeNil())
		process := &os.Process{}
		proc := &exec.Cmd{ Process: process }
		handler := &cmd.CommandStreamHandler{
			Process: proc,
			Stdout: strings.NewReader(""),
			Stderr: strings.NewReader(""),
			WaitFunc: func() error { return nil },
		}
		runner.EXPECT().Stream(cmd.New().Kubectl().With("port-forward", "-n", "ns", "deploy/dep", "1234").Cmd()).Return(handler, nil)
		runner.EXPECT().Request(req).Return("barfoo", 200, nil).Times(1)
		runner.EXPECT().Kill(process).Return(nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs with a port forward with namespace value and default port", func() {
		values := render.Values{
			"Namespace": "ns",
		}
		portFwd := validation.PortForward{
			DeploymentName: "dep",
		}
		curl := validation.Curl{
			Path:                  path,
			Host:                  "localhost",
			PortForward:           &portFwd,
			StatusCode:            200,
		}
		req, err := curl.GetHttpRequest("http://localhost:8080/path")
		Expect(err).To(BeNil())
		process := &os.Process{}
		proc := &exec.Cmd{ Process: process }
		handler := &cmd.CommandStreamHandler{
			Process: proc,
			Stdout: strings.NewReader(""),
			Stderr: strings.NewReader(""),
			WaitFunc: func() error { return nil },
		}
		runner.EXPECT().Stream(cmd.New().Kubectl().With("port-forward", "-n", "ns", "deploy/dep", "8080").Cmd()).Return(handler, nil)
		runner.EXPECT().Request(req).Return("barfoo", 200, nil).Times(1)
		runner.EXPECT().Kill(process).Return(nil).Times(1)
		err = curl.Run(ctx, values)
		Expect(err).To(BeNil())
	})
})
