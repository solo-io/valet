package two_phased_canary_test

import (
	"context"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/valet/pkg/docs"
	"github.com/solo-io/valet/pkg/step/check"
	"github.com/solo-io/valet/pkg/step/helm"
	"github.com/solo-io/valet/pkg/workflow"
	"io/ioutil"
	"os"
	"testing"
)

func TestTwoPhasedCanary(t *testing.T) {
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()
	RunSpecs(t, "Two Phased Canary Test Suite")
}

var _ = Describe("Two Phased Canary", func() {

	ctx := workflow.DefaultContext(context.TODO())

	installGloo := func() *workflow.Step {
		return &workflow.Step{
			InstallHelmChart: &helm.InstallHelmChart{
				ReleaseName: "gloo",
				ReleaseUri:  "https://storage.googleapis.com/gloo-ee-helm/charts/gloo-ee-1.3.0.tgz",
				Namespace:   "gloo-system",
				WaitForPods: true,
				Set: map[string]string{
					"license_key": "env:LICENSE_KEY",
				},
				ValuesFiles: []string{"values.yaml"},
			},
		}
	}

	gatewayProxy := func() *check.ServiceRef {
		return &check.ServiceRef{
			Namespace: "gloo-system",
			Name:      "gateway-proxy",
		}
	}

	curl := func(responseBody string) *workflow.Step {
		return &workflow.Step{
			Curl: &check.Curl{
				Service:      gatewayProxy(),
				StatusCode:   200,
				Path:         "/",
				ResponseBody: responseBody,
			},
		}
	}

	curlWithHeader := func(responseBody, headerName, headerValue string) *workflow.Step{
		step := curl(responseBody)
		step.Curl.Headers = map[string]string{
			headerName: headerValue,
		}
		return step
	}

	getWorkflow := func() *workflow.Workflow {
		return &workflow.Workflow{
			Steps: []*workflow.Step{
				installGloo(),

				// Part 1: Deploy the app
				workflow.Apply("echo.yaml").WithId("deploy-echo"),
				workflow.WaitForPods("echo").WithId("wait-1"),
				workflow.Apply("upstream.yaml").WithId("deploy-upstream"),
				workflow.Apply("vs-1.yaml").WithId("deploy-vs-1"),
				curl("version:v1"),

				// Part 2: Initial subset routing
				workflow.Apply("vs-2.yaml").WithId("deploy-vs-2"),
				curl("version:v1"),

				// Part 3: Deploy v2 with subset route
				workflow.Apply("echo-v2.yaml").WithId("deploy-echo-v2"),
				workflow.WaitForPods("echo").WithId("wait-1"),
				workflow.Apply("vs-3.yaml").WithId("deploy-vs-3"),
				curl("version:v1"),
				curlWithHeader("version:v2", "stage", "canary"),

				// Part 4: Setup weighted destinations, 0% to v2
				workflow.Apply("vs-4.yaml").WithId("deploy-vs-4"),
				curl("version:v1"),
				curlWithHeader("version:v2", "stage", "canary"),

				// Part 5: Start shift, 50% to v1 and 50% to v2
				workflow.Apply("vs-5.yaml").WithId("deploy-vs-5"),
				curl("version:v1"),
				curl("version:v2"),

				// Part 6: Finish shift, 100% to v2
				workflow.Apply("vs-6.yaml").WithId("deploy-vs-6"),
				curl("version:v2"),

				// Part 7: Decommission v1
				workflow.Delete("echo-v1.yaml").WithId("delete-echo-v1"),
				curl("version:v2"),

				// Part 8: Cleanup routes
				workflow.Apply("vs-7.yaml").WithId("deploy-vs-7"),
				curl("version:v2"),
			},
		}
	}

	It("runs", func() {
		globalConfig, err := workflow.LoadDefaultGlobalConfig(ctx.FileStore)
		Expect(err).To(BeNil())
		err = workflow.LoadEnv(globalConfig)
		Expect(err).To(BeNil())
		err = getWorkflow().Run(ctx)
		Expect(err).To(BeNil())
	})

	It("can serialize as and deserialize from yaml", func() {
		initial := getWorkflow()
		bytes, err := yaml.Marshal(initial)
		Expect(err).To(BeNil())
		err = ioutil.WriteFile("workflow.yaml", bytes, os.ModePerm)
		Expect(err).To(BeNil())
		deserialized := &workflow.Workflow{}
		err = yaml.UnmarshalStrict(bytes, deserialized, yaml.DisallowUnknownFields)
		Expect(err).To(BeNil())
		Expect(deserialized).To(Equal(initial))
	})

	It("can produce docs", func() {
		err := docs.ProcessDoc(workflow.DefaultContext(context.TODO()), "template.md", "README.md")
		Expect(err).To(BeNil())
	})
})
