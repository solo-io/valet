package check_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/pkg/api"
	mockkube "github.com/solo-io/valet/pkg/client/kube/mocks"
	"github.com/solo-io/valet/pkg/cmd"
	mockcmd "github.com/solo-io/valet/pkg/cmd/mocks"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/step/check"
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
		runner          *mockcmd.MockRunner
		kubeClient      *mockkube.MockClient
		ctx             *api.WorkflowContext
		gatewayProxySvc = &check.ServiceRef{Namespace: svcNs, Name: svcName}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mockcmd.NewMockRunner(ctrl)
		kubeClient = mockkube.NewMockClient(ctrl)
		ctx = &api.WorkflowContext{
			Runner:     runner,
			KubeClient: kubeClient,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("runs with default method", func() {
		curl := check.Curl{
			Path:       path,
			Host:       host,
			Service:    gatewayProxySvc,
			StatusCode: 200,
		}
		kubeClient.EXPECT().GetIngressAddress(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := http.NewRequest(check.DefaultMethod, "http://host/path", nil)
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs with the right default attempts during status code failure", func() {
		curl := check.Curl{
			Path:       path,
			Host:       host,
			Service:    gatewayProxySvc,
			StatusCode: 200,
			Delay:      "1ms",
		}
		kubeClient.EXPECT().GetIngressAddress(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := http.NewRequest(check.DefaultMethod, "http://host/path", nil)
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("", 503, nil).Times(10)
		err = curl.Run(ctx, nil)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal(check.UnexpectedStatusCodeError(503).Error()))
	})

	It("runs with a specific set of attempts during body failure", func() {
		curl := check.Curl{
			Path:         path,
			Host:         host,
			Service:      gatewayProxySvc,
			StatusCode:   200,
			Delay:        "1ms",
			Attempts:     1,
			ResponseBody: "foo",
		}
		kubeClient.EXPECT().GetIngressAddress(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := http.NewRequest(check.DefaultMethod, "http://host/path", nil)
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("bar", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal(check.UnexpectedResponseBodyError("bar").Error()))
	})

	It("runs checking response body", func() {
		curl := check.Curl{
			Path:         path,
			Host:         host,
			Service:      gatewayProxySvc,
			StatusCode:   200,
			ResponseBody: "bar",
		}
		kubeClient.EXPECT().GetIngressAddress(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := curl.GetHttpRequest("http://host/path")
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("bar", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs checking response body substring", func() {
		curl := check.Curl{
			Path:                  path,
			Host:                  host,
			Service:               gatewayProxySvc,
			StatusCode:            200,
			ResponseBodySubstring: "bar",
		}
		kubeClient.EXPECT().GetIngressAddress(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := curl.GetHttpRequest("http://host/path")
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("barfoo", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs with default status code 200", func() {
		curl := check.Curl{
			Path:    path,
			Host:    host,
			Service: gatewayProxySvc,
		}
		kubeClient.EXPECT().GetIngressAddress(svcName, svcNs, svcPort).Return(host, nil).Times(1)
		req, err := curl.GetHttpRequest("http://host/path")
		Expect(err).To(BeNil())
		runner.EXPECT().Request(req).Return("", 200, nil).Times(1)
		err = curl.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs with a port forward fully specified", func() {
		portFwd := check.PortForward{
			Namespace:      "ns",
			DeploymentName: "dep",
			Port:           1234,
		}
		curl := check.Curl{
			Path:        path,
			Host:        "localhost",
			PortForward: &portFwd,
			StatusCode:  200,
		}
		req, err := curl.GetHttpRequest("http://localhost:1234/path")
		Expect(err).To(BeNil())
		process := &os.Process{}
		proc := &exec.Cmd{Process: process}
		handler := &cmd.CommandStreamHandler{
			Process:  proc,
			Stdout:   strings.NewReader(""),
			Stderr:   strings.NewReader(""),
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
		portFwd := check.PortForward{
			DeploymentName: "dep",
		}
		curl := check.Curl{
			Path:        path,
			Host:        "localhost",
			PortForward: &portFwd,
			StatusCode:  200,
		}
		req, err := curl.GetHttpRequest("http://localhost:8080/path")
		Expect(err).To(BeNil())
		process := &os.Process{}
		proc := &exec.Cmd{Process: process}
		handler := &cmd.CommandStreamHandler{
			Process:  proc,
			Stdout:   strings.NewReader(""),
			Stderr:   strings.NewReader(""),
			WaitFunc: func() error { return nil },
		}
		runner.EXPECT().Stream(cmd.New().Kubectl().With("port-forward", "-n", "ns", "deploy/dep", "8080").Cmd()).Return(handler, nil)
		runner.EXPECT().Request(req).Return("barfoo", 200, nil).Times(1)
		runner.EXPECT().Kill(process).Return(nil).Times(1)
		err = curl.Run(ctx, values)
		Expect(err).To(BeNil())
	})
})
