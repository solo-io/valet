package check_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/pkg/api"
	mockkube "github.com/solo-io/valet/pkg/client/kube/mocks"
	mockcmd "github.com/solo-io/valet/pkg/cmd/mocks"
	"github.com/solo-io/valet/pkg/step/check"
)

var _ = Describe("wait_for_pods", func() {

	const (
		ns      = "wait-ns"
	)

	var (
		ctrl                *gomock.Controller
		runner              *mockcmd.MockRunner
		kubeClient          *mockkube.MockClient
		ctx                 *api.WorkflowContext
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

	It("works", func() {
		kubeClient.EXPECT().WaitUntilPodsRunning(ns).Return(nil).Times(1)
		step := check.WaitForPods{ Namespace: ns }
		err := step.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

})
