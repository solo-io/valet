package docs

import (
	"github.com/ghodss/yaml"
	"github.com/rotisserie/eris"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"
	"github.com/solo-io/valet/pkg/workflow"
	"regexp"
	"strings"
)

func ProcessDoc(ctx *api.WorkflowContext, templatePath, outputPath string) error {
	contents, err := ctx.FileStore.Load(templatePath)
	if err != nil {
		return err
	}
	re := regexp.MustCompile("(?s)\\{\\{\\%valet(.*?)\\%\\}\\}")
	matches := re.FindAllStringSubmatch(contents, -1)
	for _, match := range matches {
		fullSubstring := match[0]
		interior := match[1]
		docsRef := &DocsRef{}
		if err := yaml.UnmarshalStrict([]byte(interior), docsRef); err != nil {
			return err
		}
		docs, err := GetDocsForRef(ctx, docsRef)
		if err != nil {
			return err
		}
		contents = strings.Replace(contents, fullSubstring, docs, 1)
	}
	return ctx.FileStore.Save(outputPath, contents)
}

func GetDocsForRef(ctx *api.WorkflowContext, ref *DocsRef) (string, error) {
	deserialized := &workflow.Workflow{}
	if err := ctx.FileStore.LoadYaml(ref.Workflow, deserialized); err != nil {
		return "", err
	}
	for _, step := range deserialized.Steps {
		if ref.Step != "" && ref.Step == step.Id {
			workflowValues := deserialized.Values.MergeValues(step.Values)
			return step.Get().GetDocs(ctx, workflowValues, ref.Flags)
		}

	}
	return "", eris.Errorf("Step not found! Workflow: %s, Step: %s", ref.Workflow, ref.Step)
}

type DocsRef struct {
	Workflow string       `json:"workflow,omitempty"`
	Step     string       `json:"step,omitempty"`
	Flags    render.Flags `json:"flags,omitempty"`
}
