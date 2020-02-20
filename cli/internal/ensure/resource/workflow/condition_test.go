package workflow_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	mock_cmd "github.com/solo-io/valet/cli/internal/ensure/cmd/mocks"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"github.com/solo-io/valet/cli/internal/ensure/resource/workflow"
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
		ctrl   *gomock.Controller
		runner *mock_cmd.MockRunner
		input  render.InputParams

		ctx         = context.TODO()
		emptyErr    = errors.Errorf("")
		expectedCmd = cmd.New().Kubectl().With("get", kubeType, name, "-n", namespace, fmt.Sprintf("-o=jsonpath=%s", jsonPath)).Cmd()
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		input = render.InputParams{
			CommandRunner: runner,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("fully provided condition", func() {
		condition := &workflow.Condition{
			Name:      name,
			Namespace: namespace,
			Type:      kubeType,
			Jsonpath:  jsonPath,
			Value:     value,
			Timeout:   timeout,
			Interval:  interval,
		}

		It("works for immediately successful condition", func() {
			runner.EXPECT().Output(ctx, expectedCmd).Return(value, nil).Times(1)
			err := condition.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})

		It("works for failed condition", func() {
			runner.EXPECT().Output(ctx, expectedCmd).Return("", emptyErr).Times(1)
			err := condition.Ensure(ctx, input)
			Expect(err).To(Equal(emptyErr))
		})

		It("works for condition not met", func() {
			runner.EXPECT().Output(ctx, expectedCmd).Return(otherValue, nil).AnyTimes()
			err := condition.Ensure(ctx, input)
			Expect(err).To(Equal(workflow.ConditionNotMetError))
		})

		It("works for condition met eventually", func() {
			runner.EXPECT().Output(ctx, expectedCmd).Return(otherValue, nil).Times(3)
			runner.EXPECT().Output(ctx, expectedCmd).Return(value, nil).Times(1)
			err := condition.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})
	})

	Context("condition template rendering", func() {

		condition := &workflow.Condition{
			Name:      "{{ .Name }}",
			Namespace: namespace,
			Type:      kubeType,
			Jsonpath:  jsonPath,
			Value:     value,
			Timeout:   "{{ .Timeout }}",
			Interval:  "{{ .Interval }}",
		}

		const (
			nameKey     = "Name"
			timeoutKey  = "Timeout"
			intervalKey = "Interval"
		)

		var (
			values render.Values
		)

		BeforeEach(func() {
			values = make(map[string]string)
			values[nameKey] = name
			values[timeoutKey] = timeout
			values[intervalKey] = interval
			input.Values = values
		})

		It("works", func() {
			runner.EXPECT().Output(ctx, expectedCmd).Return(value, nil).Times(1)
			err := condition.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})

		It("returns error if missing value", func() {
			condition.Name = "{{ .MissingValue }}"
			err := condition.Ensure(ctx, input)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("MissingValue"))
		})
	})

	Context("condition default rendering", func() {

		condition := &workflow.Condition{}

		It("works", func() {
			err := input.RenderFields(condition)
			Expect(err).To(BeNil())
			Expect(condition.Timeout).To(Equal(workflow.DefaultConditionTimeout))
			Expect(condition.Interval).To(Equal(workflow.DefaultConditionInterval))
		})
	})

})
