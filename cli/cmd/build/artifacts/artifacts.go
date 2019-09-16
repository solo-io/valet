package artifacts

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
)

const (
	DefaultArtifactsFile = "artifacts.yaml"
)

func ReadArtifactsConfig(path string) (*Artifacts, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var artifacts Artifacts
	if err := yaml.Unmarshal(bytes, &artifacts); err != nil {
		return nil, err
	}

	return &artifacts, nil
}

type Artifacts struct {
	Build       Build  `json:"build,omitempty"`
	Docker      Docker `json:"docker,omitempty"`
	Helm        Helm   `json:"helm,omitempty"`
	ProductName string `json:"productName,omitempty"`
}

type Build struct {
	Go Go `json:"go,omitempty"`
}

type Go struct {
	Version  string   `json:"version,omitempty"`
	GcFlags  string   `json:"gcFlags,omitempty"`
	Binaries []Binary `json:"binaries,omitempty"`
}

type Binary struct {
	Name       string   `json:"name,omitempty"`
	Os         []string `json:"os,omitempty"`
	Entrypoint string   `json:"entrypoint,omitempty"`
	Tests      []Test   `json:"tests,omitempty"`
}

type Test struct {
	Path string `json:"path,omitempty"`
}

type Docker struct {
	Registries []string    `json:"registries,omitempty"`
	Containers []Container `json:"containers,omitempty"`
}

type Container struct {
	Name       string `json:"name,omitempty"`
	Dockerfile string `json:"dockerfile,omitempty"`
}

type Helm struct {
	Charts []Chart `json:"charts,omitempty"`
}

type Chart struct {
	Name      string     `json:"name,omitempty"`
	Directory string     `json:"directory,omitempty"`
	Generator string     `json:"generator,omitempty"`
	Manifests []Manifest `json:"manifests,omitempty"`
}

type Manifest struct {
	Name   string `json:"name,omitempty"`
	Values string `json:"values,omitempty"`
}
