package resource_test

import (
	"context"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/cmd/mocks"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
)

var _ = Describe("cert", func() {

	const (
		name      = "cert"
		namespace = "gloo-system"
		domain    = "foo.corp.solo.io"
	)

	var (
		ctx      = context.TODO()
		ctrl     *gomock.Controller
		runner   *mocks.MockCommandRunner
		command  cmd.Factory
		emptyErr = errors.Errorf("")

		TestCertResource = &resource.Cert{
			Name:      name,
			Domain:    domain,
			Namespace: namespace,
		}

		ExpectedCreateCertCommand = func(runner cmd.CommandRunner) *cmd.Command {
			return &cmd.Command{
				Name: "kubectl",
				Args: []string{
					"apply", "-f", "-",
				},
				StdIn:         TestCertResource.GetCertYaml(),
				CommandRunner: runner,
			}
		}

		ExpectedDeleteCertCommand = func(runner cmd.CommandRunner) *cmd.Command {
			return &cmd.Command{
				Name: "kubectl",
				Args: []string{
					"delete", "-f", "-",
				},
				StdIn:         TestCertResource.GetCertYaml(),
				CommandRunner: runner,
			}
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(test)
		runner = mocks.NewMockCommandRunner(ctrl)
		command = &cmd.CommandFactory{
			CommandRunner: runner,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("successfully ensures", func() {
		runner.EXPECT().Run(ctx, ExpectedCreateCertCommand(runner)).Return(nil)
		err := TestCertResource.Ensure(ctx, command)
		Expect(err).To(BeNil())
	})

	It("propagates error on ensure", func() {
		runner.EXPECT().Run(ctx, ExpectedCreateCertCommand(runner)).Return(emptyErr)
		err := TestCertResource.Ensure(ctx, command)
		Expect(err).To(Equal(emptyErr))
	})

	It("successfully tears down", func() {
		runner.EXPECT().Run(ctx, ExpectedDeleteCertCommand(runner)).Return(nil)
		err := TestCertResource.Teardown(ctx, command)
		Expect(err).To(BeNil())
	})

	It("propagates error on teardown", func() {
		runner.EXPECT().Run(ctx, ExpectedDeleteCertCommand(runner)).Return(emptyErr)
		err := TestCertResource.Teardown(ctx, command)
		Expect(err).To(Equal(emptyErr))
	})
})
