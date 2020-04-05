package gloo_petclinic_test

import (
	"context"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/config"
	"github.com/solo-io/valet/cli/options"
	"github.com/solo-io/valet/pkg/docs"
	"github.com/solo-io/valet/pkg/step/helm"
	"github.com/solo-io/valet/pkg/step/kubectl"
	"github.com/solo-io/valet/pkg/step/validation"
	"github.com/solo-io/valet/pkg/workflow"
	"io/ioutil"
	"os"
	"testing"
)

func TestPetclinic(t *testing.T) {
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()
	RunSpecs(t, "Petclinic Suite")
}

var _ = Describe("petclinic", func() {

	installGloo := func() *workflow.Step {
		return &workflow.Step{
			InstallHelmChart: &helm.InstallHelmChart{
				ReleaseName: "gloo",
				ReleaseUri:  "https://storage.googleapis.com/solo-public-helm/charts/gloo-1.3.17.tgz",
				Namespace:   "gloo-system",
				WaitForPods: true,
			},
		}
	}

	gatewayProxy := func() *validation.ServiceRef {
		return &validation.ServiceRef{
			Namespace: "gloo-system",
			Name:      "gateway-proxy",
		}
	}

	createAwsSecret := func() *workflow.Step {
		return &workflow.Step{
			CreateSecret: &kubectl.CreateSecret{
				Namespace: "gloo-system",
				Name:      "aws-creds",
				Type:      "generic",
				Entries: map[string]kubectl.SecretValue{
					"aws_access_key_id":     {EnvVar: "AWS_ACCESS_KEY_ID"},
					"aws_secret_access_key": {EnvVar: "AWS_SECRET_ACCESS_KEY"},
				},
			},
		}
	}

	initialCurl := func() *workflow.Step {
		return &workflow.Step{
			Curl: &validation.Curl{
				Service:    gatewayProxy(),
				Path:       "/",
				StatusCode: 200,
			},
		}
	}

	curlVetsForUpdate := func() *workflow.Step {
		return &workflow.Step{
			Curl: &validation.Curl{
				Service:               gatewayProxy(),
				Path:                  "/vets",
				StatusCode:            200,
				ResponseBodySubstring: "Boston",
			},
		}
	}

	curlContactPageForFix := func() *workflow.Step {
		return &workflow.Step{
			Curl: &validation.Curl{
				Service:               gatewayProxy(),
				Path:                  "/contact.html",
				StatusCode:            200,
				ResponseBodySubstring: "Enter your email",
				Attempts:              30,
			},
		}
	}

	getPetclinic := func() *workflow.Workflow {
		return &workflow.Workflow{
			Steps: []*workflow.Step{
				installGloo(),
				// Part 1: Deploy the monolith
				workflow.Apply("petclinic.yaml").WithId("deploy-monolith"),
				workflow.WaitForPods("default").WithId("wait-1"),
				workflow.Apply("vs-1.yaml").WithId("vs-1"),
				initialCurl(),
				// Part 2: Extend with a new microservice
				workflow.Apply("petclinic-vets.yaml").WithId("deploy-vets"),
				workflow.WaitForPods("default").WithId("wait-2"),
				workflow.Apply("vs-2.yaml").WithId("vs-2"),
				curlVetsForUpdate(),
				// Phase 3: AWS
				createAwsSecret().WithId("aws-creds"),
				workflow.Apply("upstream-aws.yaml").WithId("upstream-aws"),
				workflow.Apply("vs-3.yaml").WithId("vs-3"),
				curlContactPageForFix(),
			},
		}
	}

	It("runs", func() {
		globalConfig, err := config.LoadGlobalConfig(&options.Options{})
		Expect(err).To(BeNil())
		err = common.LoadEnv(globalConfig)
		Expect(err).To(BeNil())
		err = getPetclinic().Run(workflow.DefaultContext(context.TODO()))
		Expect(err).To(BeNil())
	})

	It("can serialize as and deserialize from yaml", func() {
		petclinic := getPetclinic()
		bytes, err := yaml.Marshal(petclinic)
		Expect(err).To(BeNil())
		err = ioutil.WriteFile("workflow.yaml", bytes, os.ModePerm)
		Expect(err).To(BeNil())
		deserialized := &workflow.Workflow{}
		err = yaml.UnmarshalStrict(bytes, deserialized, yaml.DisallowUnknownFields)
		Expect(err).To(BeNil())
		Expect(deserialized).To(Equal(petclinic))
	})

	It("can produce docs", func() {
		err := docs.ProcessDoc(workflow.DefaultContext(context.TODO()), "template.md", "README.md")
		Expect(err).To(BeNil())
	})
})
