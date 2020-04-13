package tests

import (
	"context"
	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/pkg/docs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/workflow"
)

type TestWorkflow struct {
	Workflow          *workflow.Workflow
	Ctx               *api.WorkflowContext
	TestDocs          bool
	TestSerialization bool
}

func (t *TestWorkflow) Setup(dir string) {
	startingDir, err := os.Getwd()
	Expect(err).To(BeNil())
	Expect(os.Chdir(filepath.Join(startingDir, dir))).To(BeNil())
	globalConfig, err := workflow.LoadDefaultGlobalConfig(t.Ctx.FileStore)
	Expect(err).To(BeNil())
	err = workflow.LoadEnv(globalConfig)
	Expect(err).To(BeNil())
	err = t.Workflow.Setup(t.Ctx)
	Expect(err).To(BeNil())
	Expect(os.Chdir(startingDir)).To(BeNil())
}

func (t *TestWorkflow) Run(dir string) {
	startingDir, err := os.Getwd()
	Expect(err).To(BeNil())
	Expect(os.Chdir(filepath.Join(startingDir, dir))).To(BeNil())
	globalConfig, err := workflow.LoadDefaultGlobalConfig(t.Ctx.FileStore)
	Expect(err).To(BeNil())
	err = workflow.LoadEnv(globalConfig)
	Expect(err).To(BeNil())

	if t.TestSerialization {
		bytes, err := yaml.Marshal(t.Workflow)
		Expect(err).To(BeNil())
		err = ioutil.WriteFile("workflow.yaml", bytes, os.ModePerm)
		Expect(err).To(BeNil())
		deserialized := &workflow.Workflow{}
		err = yaml.UnmarshalStrict(bytes, deserialized, yaml.DisallowUnknownFields)
		Expect(err).To(BeNil())
		Expect(deserialized).To(Equal(t.Workflow))
	}

	if t.TestDocs {
		err := docs.ProcessDoc(t.Ctx, "template.md", "README.md")
		Expect(err).To(BeNil())
	}

	err = t.Workflow.Run(t.Ctx)
	Expect(err).To(BeNil())
	Expect(os.Chdir(startingDir)).To(BeNil())
}

func Workflow(input *workflow.Workflow) *TestWorkflow {
	return &TestWorkflow{
		Workflow: input,
		Ctx:      workflow.DefaultContext(context.TODO()),
	}
}
