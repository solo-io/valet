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

	"github.com/solo-io/valet/cli/internal"
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

func (h *HelmChart) renderManifest(ctx context.Context, input render.InputParams, command cmd.Factory) (string, error) {
	if err := h.addHelmRepo(ctx, command); err != nil {
		return "", err
	}
	untarDir, err := h.fetchAndUntarChart(ctx, command)
	chartDir := filepath.Join(untarDir, h.ChartName)
	if err != nil {
		return "", err
	}
	cmd.Stdout().Println("Rendering manifest for application")
	helmCmd := command.Helm().Template().Namespace(h.Namespace)
	for _, set := range h.Set {
		helmCmd = helmCmd.Set(set)
	}
	for set, envVar := range h.SetEnv {
		helmCmd = helmCmd.SetEnv(set, envVar)
	}
	vals, err := input.Values.RenderValues()
	if err != nil {
		return "", err
	}
	for key, val := range vals {
		helmCmd = helmCmd.Set(fmt.Sprintf("%s=%s", key, val))
	}
	files, err := h.Files.RenderStringValues()
	if err != nil {
		return "", err
	}
	for key, file := range files {
		helmCmd = helmCmd.SetFile(fmt.Sprintf("%s=%s", key, file))
	}
	return helmCmd.Target(chartDir).Cmd().Output(ctx)
}

func (h *HelmChart) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	cmd.Stdout().Println("Preparing to uninstall %s version %s", h.ChartName, h.Version)
	if err := input.Values.RenderFields(h); err != nil {
		return err
	}
	manifest, err := h.renderManifest(ctx, input, command)
	if err != nil {
		return err
	}
	cmd.Stdout().Println("Deleting manifest")
	kubectl := command.Kubectl().DeleteStdIn(manifest).IgnoreNotFound()
	if h.Namespace != "" {
		kubectl = kubectl.Namespace(h.Namespace)
	}
	if err := kubectl.Cmd().Run(ctx); err != nil {
		return err
	}
	if h.Namespace != "" {
		if ok, err := internal.NamespaceIsActive(h.Namespace); err != nil {
			return err
		} else if ok {
			return command.Kubectl().Delete(ns).WithName(h.Namespace).Cmd().Run(ctx)
		}
	}
	return nil
}

func (h *HelmChart) addHelmRepo(ctx context.Context, command cmd.Factory) error {
	return command.Helm().AddRepo(h.RepoName, h.RepoUrl).Cmd().Run(ctx)
}

func (h *HelmChart) fetchAndUntarChart(ctx context.Context, command cmd.Factory) (string, error) {
	untarDir, err := h.getLocalDirectory()
	if err != nil {
		cmd.Stderr().Println("Error determining local directory for extracting chart: %s", err.Error())
		return "", err
	}
	if err := os.MkdirAll(untarDir, os.ModePerm); err != nil {
		cmd.Stderr().Println("Error making directory: %s", err.Error())
		return "", err
	}
	out, err := command.
		Helm().
		Fetch(h.RepoName, h.ChartName).
		Version(h.Version).
		UntarToDir(untarDir).
		Cmd().
		Output(ctx)
	if err != nil {
		cmd.Stderr().Println("Error trying to extract chart: %s", err.Error())
		cmd.Stderr().Println(out)
		return "", err
	}
	cmd.Stdout().Println("Successfully downloaded and extracted chart")
	return untarDir, nil
}

func (h *HelmChart) getLocalDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".helm", "untar", h.RepoName, h.ChartName, h.Version), nil
}

func (h *HelmChart) Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	if err := input.Values.RenderFields(h); err != nil {
		return nil, err
	}
	url := h.getChartUrl()
	values, err := h.computeValues(ctx, input)
	if err != nil {
		return nil, err
	}
	manifests, err := helmchart.RenderManifests(ctx,
		url,
		values,
		h.ChartName,
		h.Namespace,
		"")
	if err != nil {
		return nil, err
	}
	return manifests.ResourceList()
}

func (h *HelmChart) getChartUrl() string {
	filename := fmt.Sprintf("%s-%s.tgz", h.ChartName, h.Version)
	if strings.HasSuffix(h.RepoUrl, "/") {
		return fmt.Sprintf("%s%s/%s", h.RepoUrl, "charts", filename)
	}
	return fmt.Sprintf("%s/%s/%s", h.RepoUrl, "charts", filename)
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
