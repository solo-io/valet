package application_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("helm chart", func() {

	const (
		helmChartRepoUrl  = "https://storage.googleapis.com/sm-marketplace-helm/"
		helmChartRepoName = "smh"
		helmChartName     = "sm-marketplace"
		helmChartVersion  = "0.3.4"

		registryName       = "test-registry"
		registryPath       = "test"
		namespace          = "test-namespace"
		valuesFilePath     = "test/files/helm/example-values.yaml"
		valuesRegistryPath = "files/helm/example-values.yaml"
		setFilePath        = "test/files/helm/example-set-file.txt"
		setRegistryPath    = "files/helm/example-set-file.txt"

		setKey = "namespace.create"
	)

	var (
		ctx          = context.TODO()
		emptyInput   = render.InputParams{}
		testRegistry = render.DirectoryRegistry{
			WorkingDirectory: registryPath,
		}
		values render.Values
		input  render.InputParams
	)

	getHelmChart := func() *application.HelmChart {
		return &application.HelmChart{
			RepoUrl:   helmChartRepoUrl,
			RepoName:  helmChartRepoName,
			ChartName: helmChartName,
			Version:   helmChartVersion,
		}
	}

	namespaceFilter := func(resource *unstructured.Unstructured) bool {
		return resource.GetKind() != "Namespace"
	}

	inTestNamespaceFilter := func(resource *unstructured.Unstructured) bool {
		return resource.GetNamespace() != namespace
	}

	BeforeEach(func() {
		values = render.Values{}
		input = render.InputParams{Values: values}
		input.SetRegistry(registryName, &testRegistry)
	})

	Context("render", func() {

		It("works for a basic helm chart", func() {
			helmChart := getHelmChart()
			resources, err := helmChart.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(22))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(0))
			inTestNamespace := resources.Filter(inTestNamespaceFilter)
			Expect(len(inTestNamespace)).To(Equal(0))
		})

		It("returns error for a fake helm repo", func() {
			helmChart := getHelmChart()
			helmChart.RepoUrl = "fake/url.com"
			helmChart.RepoName = "fake-repo"
			_, err := helmChart.Render(ctx, emptyInput)
			Expect(err).NotTo(BeNil())
		})

		It("returns error for a fake chart name", func() {
			helmChart := getHelmChart()
			helmChart.ChartName = "fake-chart"
			_, err := helmChart.Render(ctx, emptyInput)
			Expect(err).NotTo(BeNil())
		})

		It("returns error for an invalid chart version", func() {
			helmChart := getHelmChart()
			helmChart.Version = "0.1.100"
			_, err := helmChart.Render(ctx, emptyInput)
			Expect(err).NotTo(BeNil())
		})

		It("works for namespacing resources", func() {
			helmChart := getHelmChart()
			helmChart.Namespace = namespace
			resources, err := helmChart.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(22))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(0))
			inTestNamespace := resources.Filter(inTestNamespaceFilter)
			Expect(len(inTestNamespace)).To(Equal(14))
		})

		It("works for namespacing resources via values", func() {
			helmChart := getHelmChart()
			values[render.NamespaceKey] = namespace
			resources, err := helmChart.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(22))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(0))
			inTestNamespace := resources.Filter(inTestNamespaceFilter)
			Expect(len(inTestNamespace)).To(Equal(14))
		})

		It("works for supplying version via values", func() {
			helmChart := getHelmChart()
			helmChart.Version = ""
			values[render.VersionKey] = helmChartVersion
			resources, err := helmChart.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(22))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(0))
			inTestNamespace := resources.Filter(inTestNamespaceFilter)
			Expect(len(inTestNamespace)).To(Equal(0))
		})

		It("works for supplying helm values via set", func() {
			helmChart := getHelmChart()
			helmChart.Set = []string{"namespace.create=true"}
			resources, err := helmChart.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(23))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(1))
		})

		It("works for supplying helm values via set env", func() {
			envVar := "TEST_HELM_VALUE"
			Expect(os.Setenv(envVar, "true")).To(BeNil())
			helmChart := getHelmChart()
			helmChart.SetEnv = map[string]string{setKey: envVar}
			resources, err := helmChart.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(23))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(1))
		})

		It("works for supplying helm values via values files with default registry", func() {
			helmChart := getHelmChart()
			helmChart.ValuesFiles = []string{valuesFilePath}
			resources, err := helmChart.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(23))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(1))
		})

		It("works for supplying helm values via values files with test registry", func() {
			helmChart := getHelmChart()
			helmChart.ValuesFiles = []string{valuesRegistryPath}
			helmChart.RegistryName = registryName
			resources, err := helmChart.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(23))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(1))
		})

		It("works for supplying helm values via set file with default registry", func() {
			helmChart := getHelmChart()
			helmChart.Files = render.Values{setKey: setFilePath}
			resources, err := helmChart.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(23))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(1))
		})

		It("works for supplying helm values via set file with test registry", func() {
			helmChart := getHelmChart()
			helmChart.Files = render.Values{setKey: setRegistryPath}
			helmChart.RegistryName = registryName
			resources, err := helmChart.Render(ctx, input)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(23))
			namespaces := resources.Filter(namespaceFilter)
			Expect(len(namespaces)).To(Equal(1))
		})
	})
})
