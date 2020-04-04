package gloo_petclinic_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/valet/cli/cmd/common"
	"github.com/solo-io/valet/cli/cmd/config"
	"github.com/solo-io/valet/cli/options"
	"github.com/solo-io/valet/pkg/step/helm"
	"github.com/solo-io/valet/pkg/step/kubectl"
	"github.com/solo-io/valet/pkg/step/validation"
	"github.com/solo-io/valet/pkg/workflow"
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

	installGloo := &workflow.Step{
		InstallHelmChart: &helm.InstallHelmChart{
			ReleaseName: "gloo",
			ReleaseUri:  "https://storage.googleapis.com/solo-public-helm/charts/gloo-1.3.17.tgz",
			Namespace:   "gloo-system",
			WaitForPods: true,
		},
	}

	gatewayProxy := &validation.ServiceRef{
		Namespace: "gloo-system",
		Name:      "gateway-proxy",
	}

	createAwsSecret := &workflow.Step{
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

	initialCurl := &workflow.Step{
		Curl: &validation.Curl{
			Service:    gatewayProxy,
			Path:       "/",
			StatusCode: 200,
		},
	}

	curlVetsForUpdate := &workflow.Step{
		Curl: &validation.Curl{
			Service:               gatewayProxy,
			Path:                  "/vets",
			StatusCode:            200,
			ResponseBodySubstring: "Boston",
		},
	}

	curlContactPageForFix := &workflow.Step{
		Curl: &validation.Curl{
			Service:               gatewayProxy,
			Path:                  "/contact.html",
			StatusCode:            200,
			ResponseBodySubstring: "Enter your email",
			Attempts:              30,
		},
	}

	It("runs", func() {
		globalConfig, err := config.LoadGlobalConfig(&options.Options{})
		Expect(err).To(BeNil())
		err = common.LoadEnv(globalConfig)
		Expect(err).To(BeNil())
		petclinic := &workflow.Workflow{
			Steps: []*workflow.Step{
				installGloo,
				// Part 1: Deploy the monolith
				workflow.Apply("https://raw.githubusercontent.com/sololabs/demos/b523571c66057a5591bce22ad896729f1fee662b/petclinic_demo/petclinic.yaml"),
				workflow.Apply("https://raw.githubusercontent.com/sololabs/demos/b523571c66057a5591bce22ad896729f1fee662b/petclinic_demo/petclinic-db.yaml"),
				workflow.WaitForPods("default"),
				workflow.Apply("resources/petclinic/vs-1.yaml"),
				initialCurl,
				// Part 2: Extend with a new microservice
				workflow.Apply("https://raw.githubusercontent.com/sololabs/demos/b523571c66057a5591bce22ad896729f1fee662b/petclinic_demo/petclinic-vets.yaml"),
				workflow.WaitForPods("default"),
				workflow.Apply("resources/petclinic/vs-2.yaml"),
				curlVetsForUpdate,
				// Phase 3: AWS
				createAwsSecret,
				workflow.Apply("resources/petclinic/upstream-aws.yaml"),
				workflow.Apply("resources/petclinic/vs-3.yaml"),
				curlContactPageForFix,
			},
		}
		err = petclinic.Run(workflow.DefaultContext(context.TODO()))
		Expect(err).To(BeNil())
	})
})
