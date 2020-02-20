package workflow

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/go-utils/installutils/helminstall"
)

type Helm3Deploy struct {
	ReleaseName string
	ReleaseUri  string
	Namespace   string
	ValuesFiles []string
}

func (h *Helm3Deploy) Ensure(ctx context.Context, inputs render.InputParams) error {
	inst := helminstall.MustInstaller()
	conf := helminstall.InstallerConfig{
		CreateNamespace:  true,
		InstallNamespace: h.Namespace,
		ReleaseName:      h.ReleaseName,
		ReleaseUri:       h.ReleaseUri,
		ValuesFiles:      h.ValuesFiles,
	}
	return inst.Install(&conf)
}

func (h Helm3Deploy) Teardown(ctx context.Context, inputs render.InputParams) error {
	panic("implement me")
}
