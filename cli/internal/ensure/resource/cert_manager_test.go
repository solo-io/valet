package resource_test

import (
	"context"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/cmd/mocks"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
)

var _ = Describe("cert_manager", func() {

	var (
		ctx      = context.TODO()
		ctrl     *gomock.Controller
		runner   *mocks.MockCommandRunner
		command  cmd.Factory

		TestCertManagerResource = &resource.CertManager{}

		ExpectedEnsureManifestCommand = func(runner cmd.CommandRunner, path string) *cmd.Command {
			return &cmd.Command{
				Name: "kubectl",
				Args: []string{
					"apply", "-f", path,
				},
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

	PIt("successfully ensures", func() {
		runner.EXPECT().Run(ctx, ExpectedEnsureManifestCommand(runner, resource.CertManagerManifest)).Return(nil)
		//runner.EXPECT().Output(ctx, ExpectedEnsureManifestCommand(runner, resource.AwsSecret(resource.CertManagerNamespace, resource.CertManagerAwsSecretName))).Return(nil)

		err := TestCertManagerResource.Ensure(ctx, command)
		Expect(err).To(BeNil())
	})
})
