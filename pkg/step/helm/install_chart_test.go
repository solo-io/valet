package helm_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/valet/pkg/api"
	mock_helm "github.com/solo-io/valet/pkg/client/helm/mocks"
	mock_kube "github.com/solo-io/valet/pkg/client/kube/mocks"
	mock_cmd "github.com/solo-io/valet/pkg/cmd/mocks"
	"github.com/solo-io/valet/pkg/step/helm"
)

var _ = Describe("install_chart", func() {

	const (
		ns      = "release-ns"
		release = "release"
		uri     = "release-helm-chart.tgz"
	)

	var (
		ctrl                *gomock.Controller
		runner              *mock_cmd.MockRunner
		kubeClient          *mock_kube.MockClient
		helmClient          *mock_helm.MockClient
		ctx                 *api.WorkflowContext
		getInstallChartStep = func() *helm.InstallHelmChart {
			return &helm.InstallHelmChart{
				ReleaseName: release,
				ReleaseUri:  uri,
				Namespace:   ns,
			}
		}
		getInstallerConfig = func() *helminstall.InstallerConfig{
			return &helminstall.InstallerConfig{
				InstallNamespace: ns,
				ReleaseUri:       uri,
				ReleaseName:      release,
				CreateNamespace:  true,
				ExtraValues:      make(map[string]interface{}),
			}
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		helmClient = mock_helm.NewMockClient(ctrl)
		kubeClient = mock_kube.NewMockClient(ctrl)
		ctx = &api.WorkflowContext{
			Runner:     runner,
			HelmClient: helmClient,
			KubeClient: kubeClient,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("simple case", func() {
		It("runs", func() {
			conf := getInstallerConfig()
			helmClient.EXPECT().Install(conf).Return(nil).Times(1)
			err := getInstallChartStep().Run(ctx, nil)
			Expect(err).To(BeNil())
		})

		It("has the right description", func() {
			Expect(getInstallChartStep().GetDescription(nil, nil)).Should(Equal("Deploying helm chart with release name release into namespace release-ns using chart uri release-helm-chart.tgz using default values"))
		})
	})

	Context("waiting for pods", func() {
		It("works", func() {
			conf := getInstallerConfig()
			helmClient.EXPECT().Install(conf).Return(nil).Times(1)
			kubeClient.EXPECT().WaitUntilPodsRunning(ns).Return(nil).Times(1)
			installChartStep := getInstallChartStep()
			installChartStep.WaitForPods = true
			err := installChartStep.Run(ctx, nil)
			Expect(err).To(BeNil())
		})
	})

})
