package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

var (
	_ Renderable = new(HelmChart)
)

type HelmChart struct {
	RegistryName string   `yaml:"registryName" valet:"default=default"`
	RepoUrl      string   `yaml:"repoUrl"`
	ChartName    string   `yaml:"chartName"`
	RepoName     string   `yaml:"repoName"`
	Version      string   `yaml:"version" valet:"key=Version"`
	Namespace    string   `yaml:"namespace" valet:"key=Namespace"`
	ValuesFiles  []string `yaml:"valuesFiles"`
	/*
		These values allow you to perform values operations before setting them as helm values.
		You can use any operations and keywords that general `values` support.
		More information on `valet` values can be found at https://github.com/solo-io/valet/tree/master/cli/internal/ensure/resource/render

	*/
	Set render.Values `yaml:"set"`

}

func (h *HelmChart) addHelmRepo(ctx context.Context, input render.InputParams) error {
	cmd.Stdout().Println("Adding helm repo %s %s", h.RepoName, h.RepoUrl)
	if err := input.Runner().Run(ctx, cmd.New().Helm().AddRepo(h.RepoName, h.RepoUrl).Cmd()); err != nil {
		return err
	}
	cmd.Stdout().Println("Running helm repo update")
	return input.Runner().Run(ctx, cmd.New().Helm().With("repo", "update").Cmd())
}

func (h *HelmChart) fetchChart(ctx context.Context, input render.InputParams) (string, error) {
	downloadDir, err := h.getLocalDirectory()
	if err != nil {
		return "", err
	}
	downloadPath := h.getDownloadPath(downloadDir)
	if exists, err := fileExists(downloadPath); err == nil && exists {
		cmd.Stdout().Println("Chart already downloaded to %s", downloadPath)
		return downloadPath, nil
	}
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		cmd.Stderr().Println("Error making directory: %s", err.Error())
		return "", err
	}
	if err := h.addHelmRepo(ctx, input); err != nil {
		return "", err
	}
	command := cmd.New().
		Helm().
		Fetch(h.RepoName, h.ChartName).
		Version(h.Version).
		With("-d", downloadDir).
		Cmd()
	out, err := input.Runner().Output(ctx, command)
	if err != nil {
		cmd.Stderr().Println("Error trying to extract chart: %s", err.Error())
		cmd.Stderr().Println(out)
		return "", err
	}
	cmd.Stdout().Println("Successfully downloaded chart %s", downloadPath)
	return downloadPath, nil
}

func fileExists(path string) (bool, error) {
	file, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	} else if err != nil {
		return false, nil
	}
	if file.IsDir() {
		return false, errors.Errorf("Unexpected directory %s", path)
	}
	return true, nil
}

func (h *HelmChart) getDownloadPath(downloadDir string) string {
	return filepath.Join(downloadDir, fmt.Sprintf("%s-%s.tgz", h.ChartName, h.Version))
}

func (h *HelmChart) getLocalDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		cmd.Stderr().Println("Error determining local directory for extracting chart: %s", err.Error())
		return "", err
	}
	return filepath.Join(home, ".helm", "cache", "valet", h.RepoName), nil
}

func (h *HelmChart) Render(ctx context.Context, input render.InputParams) (kuberesource.UnstructuredResources, error) {
	if err := input.RenderFields(h); err != nil {
		return nil, err
	}
	downloadPath, err := h.fetchChart(ctx, input)
	if err != nil {
		return nil, err
	}
	values, err := h.computeValues(ctx, input)
	if err != nil {
		return nil, err
	}
	cmd.Stdout().Println("Successfully computed helm chart values")
	manifests, err := helmchart.RenderManifests(ctx,
		downloadPath,
		values,
		h.ChartName,
		h.Namespace,
		"")
	if err != nil {
		return nil, err
	}
	cmd.Stdout().Println("Successfully rendered helm chart")
	return manifests.ResourceList()
}

func (h *HelmChart) getParams(input render.InputParams) (map[string]string, error) {
	params := make(map[string]string)

	// Render the values here as separate from other values, but use the values engine.
	// Create a placeholder values combination so set keys don't get lost.
	tempValues := input.MergeValues(h.Set).Values
	values, err := tempValues.RenderStringValues(input.Runner())
	if err != nil {
		return nil, err
	}
	for key := range h.Set {
		params[key] = values[key]
	}

	return params, nil
}

func (h *HelmChart) computeValues(ctx context.Context, input render.InputParams) (string, error) {
	params, err := h.getParams(input)
	if err != nil {
		return "", err
	}
	values, err := render.ConvertParamsToNestedMap(params)
	if err != nil {
		return "", err
	}
	for _, valuesFile := range h.ValuesFiles {
		valuesYaml, err := input.LoadFile(h.RegistryName, valuesFile)
		if err != nil {
			return "", err
		}
		v, err := render.ConvertYamlStringToNestedMap(valuesYaml)
		if err != nil {
			return "", err
		}
		values = render.CoalesceValuesMap(ctx, values, v)
	}
	return render.ConvertNestedMapToYaml(values)
}
