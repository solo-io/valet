package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	RegistryName string            `yaml:"registryName" valet:"default=default"`
	RepoUrl      string            `yaml:"repoUrl"`
	ChartName    string            `yaml:"chartName"`
	RepoName     string            `yaml:"repoName"`
	Version      string            `yaml:"version" valet:"key=Version"`
	Namespace    string            `yaml:"namespace" valet:"key=Namespace"`
	Set          []string          `yaml:"set"`
	SetEnv       map[string]string `yaml:"setEnv"`
	ValuesFiles  []string          `yaml:"valuesFiles"`
	Files        render.Values     `yaml:"files"`
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
		cmd.Stderr().Println("Error determining local directory for extracting chart: %s", err.Error())
		return "", err
	}
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		cmd.Stderr().Println("Error making directory: %s", err.Error())
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
	downloadPath := filepath.Join(downloadDir, fmt.Sprintf("%s-%s.tgz", h.ChartName, h.Version))
	cmd.Stdout().Println("Successfully downloaded chart %s", downloadPath)
	return downloadPath, nil
}

func (h *HelmChart) getLocalDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".helm", "cache", "valet", h.RepoName), nil
}

func (h *HelmChart) Render(ctx context.Context, input render.InputParams) (kuberesource.UnstructuredResources, error) {
	if err := input.RenderFields(h); err != nil {
		return nil, err
	}
	if err := h.addHelmRepo(ctx, input); err != nil {
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
	for _, set := range h.Set {
		parts := strings.Split(set, "=")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid format (must be A=B): %s", set)
		}
		params[parts[0]] = parts[1]
	}
	for param, envVar := range h.SetEnv {
		val := os.Getenv(envVar)
		if val == "" {
			return nil, errors.Errorf("Environment variable %s not set", envVar)
		}
		params[param] = val
	}
	for param, file := range h.Files {
		contents, err := input.LoadFile(h.RegistryName, file)
		if err != nil {
			return nil, err
		}
		params[param] = contents
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
