package helm

import (
	"fmt"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/render"
)

var _ api.Step = new(InstallHelmChart)

type InstallHelmChart struct {
	ReleaseName string   `json:"releaseName,omitempty"`
	ReleaseUri  string   `json:"releaseUri,omitempty"`
	Namespace   string   `json:"namespace,omitempty"`
	ValuesFiles []string `json:"valuesFiles,omitempty"`
	/*
		These values allow you to perform values operations before setting them as helm values.
		You can use any operations and keywords that general `values` support.
		More information on `valet` values can be found at TODO
	*/
	Set         render.Values `json:"set,omitempty"`
	WaitForPods bool          `json:"waitForPods,omitempty"`
}

func (i *InstallHelmChart) GetDescription(_ *api.WorkflowContext, _ render.Values) (string, error) {
	str := fmt.Sprintf("Deploying helm chart with release name %s into namespace %s using chart uri %s", i.ReleaseName, i.Namespace, i.ReleaseUri)
	if len(i.ValuesFiles) == 0 && len(i.Set) == 0 {
		str = str + " using default values"
	}
	if len(i.ValuesFiles) > 0 {
		str = str + fmt.Sprintf(", using values files %v", i.ValuesFiles)
	}
	if len(i.Set) > 0 {
		str = str + fmt.Sprintf(", using extra values %v", i.Set)
	}
	if i.WaitForPods {
		str = str + " and waiting for the pods to be ready"
	}
	return str, nil
}

func (i *InstallHelmChart) Run(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(i, ctx.Runner); err != nil {
		return err
	}
	extraVals, err := i.Set.Render(ctx.Runner)
	if err != nil {
		return err
	}
	conf := helminstall.InstallerConfig{
		CreateNamespace:  true,
		InstallNamespace: i.Namespace,
		ReleaseName:      i.ReleaseName,
		ReleaseUri:       i.ReleaseUri,
		ValuesFiles:      i.ValuesFiles,
		ExtraValues:      extraVals,
	}
	if err := ctx.HelmClient.Install(&conf); err != nil {
		if !eris.Is(err, helminstall.ReleaseAlreadyInstalledErr(i.ReleaseName, i.Namespace)) {
			return err
		}
	}
	if !i.WaitForPods {
		return nil
	}
	return ctx.KubeClient.WaitUntilPodsRunning(i.Namespace)
}

func (i *InstallHelmChart) GetDocs(ctx *api.WorkflowContext, options api.DocsOptions) (string, error) {
	panic("implement me")
}
