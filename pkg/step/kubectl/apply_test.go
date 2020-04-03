package kubectl_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	mock_cmd "github.com/solo-io/valet/pkg/cmd/mocks"
	"github.com/solo-io/valet/pkg/step/kubectl"
)

var _ = Describe("apply", func() {

	const (
		path = "test-path"
	)

	var (
		ctrl       *gomock.Controller
		runner     *mock_cmd.MockRunner
		ctx        *api.WorkflowContext
		expectedCmd = cmd.New().Kubectl().With("apply", "-f", path).Cmd()
		apply       = kubectl.Apply{
			Path: path,
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		ctx = &api.WorkflowContext{
			Runner:     runner,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("runs", func() {
		actualCmd := apply.GetCmd()
		Expect(expectedCmd).To(Equal(actualCmd))
		runner.EXPECT().Run(actualCmd).Return(nil).Times(1)
		err := apply.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("returns the right description", func() {
		Expect(apply.GetDescription()).To(Equal("Running command: kubectl apply -f test-path"))
	})
})
