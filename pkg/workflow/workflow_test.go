package workflow_test

import (
	"github.com/golang/mock/gomock"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/valet/pkg/api"
	mock_helm "github.com/solo-io/valet/pkg/client/helm/mocks"
	mock_kube "github.com/solo-io/valet/pkg/client/kube/mocks"
	"github.com/solo-io/valet/pkg/cmd"
	mock_cmd "github.com/solo-io/valet/pkg/cmd/mocks"
	mock_render "github.com/solo-io/valet/pkg/render/mocks"
	"github.com/solo-io/valet/pkg/step/helm"
	"github.com/solo-io/valet/pkg/step/kubectl"
	"github.com/solo-io/valet/pkg/workflow"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("workflow", func() {

	var (
		ctrl   *gomock.Controller
		runner *mock_cmd.MockRunner
		fileStore  *mock_render.MockFileStore
		kubeClient   *mock_kube.MockClient
		helmClient   *mock_helm.MockClient
		ctx    *api.WorkflowContext
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		fileStore = mock_render.NewMockFileStore(ctrl)
		helmClient = mock_helm.NewMockClient(ctrl)
		kubeClient = mock_kube.NewMockClient(ctrl)
		ctx = &api.WorkflowContext{
			Runner:     runner,
			FileStore:  fileStore,
			HelmClient: helmClient,
			KubeClient: kubeClient,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	getWorkflow := func(step workflow.Step) *workflow.Workflow {
		return &workflow.Workflow{
			Steps: []workflow.Step{step},
		}
	}

	Context("apply", func() {
		const (
			path = "test-path"
		)
		var (
			expectedCmd = cmd.New().Kubectl().With("apply", "-f", path).Cmd()
			apply       = kubectl.Apply{
				Path: path,
			}
			step     = workflow.Step{Apply: &apply}
			workflow = getWorkflow(step)
		)
		It("works", func() {
			actualCmd := apply.GetCmd()
			Expect(expectedCmd).To(Equal(actualCmd))
			runner.EXPECT().Run(actualCmd).Return(nil).Times(1)
			err := workflow.Run(ctx)
			Expect(err).To(BeNil())
		})
	})

	Context("helm chart", func() {
		const (
			ns = "release-ns"
			release = "release"
			uri = "release-helm-chart.tgz"
		)
		var (
			getInstallChartStep = func() *helm.InstallHelmChart {
				return &helm.InstallHelmChart{
					ReleaseName: release,
					ReleaseUri: uri,
					Namespace: ns,
				}
			}
			getWorkflowStep = func(chart *helm.InstallHelmChart) workflow.Step {
				return workflow.Step{InstallHelmChart: chart}
			}
		)
		It("works in simple case", func() {
			conf := helminstall.InstallerConfig{
				InstallNamespace: ns,
				ReleaseUri: uri,
				ReleaseName: release,
				CreateNamespace: true,
				ExtraValues: make(map[string]interface{}),
			}
			helmClient.EXPECT().Install(&conf).Return(nil).Times(1)
			workflow := getWorkflow(getWorkflowStep(getInstallChartStep()))
			err := workflow.Run(ctx)
			Expect(err).To(BeNil())
		})

		It("works when waiting for pods", func() {
			conf := helminstall.InstallerConfig{
				InstallNamespace: ns,
				ReleaseUri: uri,
				ReleaseName: release,
				CreateNamespace: true,
				ExtraValues: make(map[string]interface{}),
			}
			helmClient.EXPECT().Install(&conf).Return(nil).Times(1)
			kubeClient.EXPECT().WaitUntilPodsRunning(ns).Return(nil).Times(1)
			installChartStep := getInstallChartStep()
			installChartStep.WaitForPods = true
			workflow := getWorkflow(getWorkflowStep(installChartStep))
			err := workflow.Run(ctx)
			Expect(err).To(BeNil())
		})
	})

})
