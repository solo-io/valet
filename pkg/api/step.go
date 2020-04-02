package api

import (
	"context"
	"github.com/solo-io/valet/pkg/client/helm"
	"github.com/solo-io/valet/pkg/client/kube"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

type Step interface {
	GetDescription() string
	Run(ctx *WorkflowContext, values render.Values) error
	GetDocs(ctx *WorkflowContext, options DocsOptions) (string, error)
}

type WorkflowContext struct {
	Ctx context.Context
	//Logger
	//SharedState
	Runner     cmd.Runner
	FileStore  render.FileStore
	HelmClient helm.Client
	KubeClient kube.Client
}

type DocsOptions map[string]string
