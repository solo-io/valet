package workflow

import (
	"context"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/go-utils/installutils/helminstall"
)

type Helm3Deploy struct {
	ReleaseName string   `json:"releaseName"`
	ReleaseUri  string   `json:"releaseUri"`
	Namespace   string   `json:"namespace"`
	ValuesFiles []string `json:"valuesFiles"`
	/*
		These values allow you to perform values operations before setting them as helm values.
		You can use any operations and keywords that general `values` support.
		More information on `valet` values can be found at https://github.com/solo-io/valet/tree/master/cli/internal/ensure/resource/render

	*/
	Set render.Values `json:"set"`
}

func (h *Helm3Deploy) Ensure(ctx context.Context, inputs render.InputParams) error {
	cmd.Stdout().Println("Running helm install to namespace %s for release %s with uri %s", h.Namespace, h.ReleaseName, h.ReleaseUri)
	extraVals, err := h.Set.Render(inputs.CommandRunner)
	if err != nil {
		return err
	}
	inst := helminstall.MustInstaller()
	conf := helminstall.InstallerConfig{
		CreateNamespace:  true,
		InstallNamespace: h.Namespace,
		ReleaseName:      h.ReleaseName,
		ReleaseUri:       h.ReleaseUri,
		ValuesFiles:      h.ValuesFiles,
		ExtraValues:      extraVals,
	}
	if err := inst.Install(&conf); err != nil {
		return err
	}
	return internal.WaitUntilPodsRunning(h.Namespace)
}

func (h *Helm3Deploy) Teardown(ctx context.Context, inputs render.InputParams) error {
	cmd.Stdout().Println("Running helm uninstall on release %s in namespace %s", h.ReleaseName, h.Namespace)
	helmClient := helminstall.DefaultHelmClient()
	uninstaller, err := helmClient.NewUninstall("", "", h.Namespace)
	if err != nil {
		return err
	}
	_, err = uninstaller.Run(h.ReleaseName)
	return err
}
