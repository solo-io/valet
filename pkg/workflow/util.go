package workflow

import (
	"context"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/client/helm"
	"github.com/solo-io/valet/pkg/client/kube"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
)

func DefaultContext(ctx context.Context) *api.WorkflowContext {
	return &api.WorkflowContext{
		Ctx:        ctx,
		Runner:     cmd.DefaultCommandRunner(),
		FileStore:  render.NewFileStore(),
		HelmClient: helm.NewClient(),
		KubeClient: kube.NewClient(),
	}
}

