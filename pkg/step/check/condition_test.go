package check_test

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	mock_cmd "github.com/solo-io/valet/pkg/cmd/mocks"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/step/check"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	errors "github.com/rotisserie/eris"
)

var _ = Describe("condition", func() {
	const (
		name      = "test-name"
		namespace = "test-namespace"
		kubeType  = "kube-type"
		jsonPath  = "json-path"
		value     = "value"
		timeout   = "10ms"
		interval  = "1ms"

		otherValue = "other-value"
	)

	var (
		ctrl        *gomock.Controller
		runner      *mock_cmd.MockRunner
		ctx         *api.WorkflowContext
		emptyErr    = errors.Errorf("")
		expectedCmd = cmd.New().Kubectl().With("get", kubeType, name, "-n", namespace, fmt.Sprintf("-o=jsonpath=%s", jsonPath)).Cmd()
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		ctx = &api.WorkflowContext{
			Runner: runner,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("fully provided condition", func() {
		condition := &check.Condition{
			Name:      name,
			Namespace: namespace,
			Type:      kubeType,
			Jsonpath:  jsonPath,
			Value:     value,
			Timeout:   timeout,
			Interval:  interval,
		}

		It("works for immediately successful condition", func() {
			runner.EXPECT().Output(expectedCmd).Return(value, nil).Times(1)
			err := condition.Run(ctx, nil)
			Expect(err).To(BeNil())
		})

		It("works for failed condition", func() {
			runner.EXPECT().Output(expectedCmd).Return("", emptyErr).AnyTimes()
			err := condition.Run(ctx, nil)
			Expect(err).To(Equal(check.ConditionNotMetError))
		})

		It("works for condition not met", func() {
			runner.EXPECT().Output(expectedCmd).Return(otherValue, nil).AnyTimes()
			err := condition.Run(ctx, nil)
			Expect(err).To(Equal(check.ConditionNotMetError))
		})

		It("works for condition met eventually", func() {
			runner.EXPECT().Output(expectedCmd).Return(otherValue, nil).Times(3)
			runner.EXPECT().Output(expectedCmd).Return(value, nil).Times(1)
			err := condition.Run(ctx, nil)
			Expect(err).To(BeNil())
		})
	})

	Context("condition template rendering", func() {

		const (
			nameKey     = "Name"
			timeoutKey  = "Timeout"
			intervalKey = "Interval"
		)

		var (
			values render.Values
			condition *check.Condition
		)

		BeforeEach(func() {
			values = make(map[string]string)
			values[nameKey] = name
			values[timeoutKey] = timeout
			values[intervalKey] = interval
			condition = &check.Condition{
				Name:      "{{ .Name }}",
				Namespace: namespace,
				Type:      kubeType,
				Jsonpath:  jsonPath,
				Value:     value,
				Timeout:   "{{ .Timeout }}",
				Interval:  "{{ .Interval }}",
			}
		})

		It("works", func() {
			runner.EXPECT().Output(expectedCmd).Return(value, nil).Times(1)
			err := condition.Run(ctx, values)
			Expect(err).To(BeNil())
		})

		It("returns error if missing value", func() {
			condition.Name = "{{ .MissingValue }}"
			err := condition.Run(ctx, values)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("MissingValue"))
		})
	})

	Context("condition default rendering", func() {

		condition := &check.Condition{}

		It("works", func() {
			values := new(render.Values)
			err := values.RenderFields(condition, ctx.Runner)
			Expect(err).To(BeNil())
			Expect(condition.Timeout).To(Equal(check.DefaultConditionTimeout))
			Expect(condition.Interval).To(Equal(check.DefaultConditionInterval))
		})
	})

})
