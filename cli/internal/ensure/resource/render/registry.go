package render

import (
	"bytes"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/workflow"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	DefaultRegistry = "default"
)

var (
	InvalidApplicationRefError = errors.Errorf("Invalid application ref")
	InvalidWorkflowRefError = errors.Errorf("Invalid application ref")
	UnknownRegistryError = func(name string) error {
		return errors.Errorf("Unknown registry %s", name)
	}
)

type Registry interface {
	LoadFile(path string) (string, error)
}

type LocalRegistry struct {
	WorkingDirectory string
}

func (l *LocalRegistry) resolvePath(path string) string {
	if l.isValidUrl(path) || l.WorkingDirectory == "" {
		return path
	}
	return filepath.Join(l.WorkingDirectory, path)
}

func (l *LocalRegistry) LoadFile(path string) (string, error) {
	return l.loadFile(l.resolvePath(path))
}




func (l *LocalRegistry) LoadWorkflow(ref *workflow.Ref) (*workflow.Workflow, error) {
	if ref.Path != "" {
		return l.loadWorkflow(ref.Path)
	}
	return nil, InvalidWorkflowRefError
}

func (l *LocalRegistry) loadWorkflow(path string) (*workflow.Workflow, error) {
	var w workflow.Workflow
	b, err := l.loadBytes(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalStrict(b, &w); err != nil {
		cmd.Stderr().Println("Failed to unmarshal file '%s': %s", path, err.Error())
		return nil, err
	}
	return &w, nil
}

func (l *LocalRegistry) loadFile(path string) (string, error) {
	b, err := l.loadBytes(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (l *LocalRegistry) loadBytes(path string) ([]byte, error) {
	if l.isValidUrl(path) {
		contents, err := LoadBytesFromUrl(path)
		if err == nil {
			return contents, nil
		}
	}

	osClient := osutils.NewOsClient()
	expandedPath := os.ExpandEnv(path)
	contents, err := osClient.ReadFile(expandedPath)
	if err != nil {
		cmd.Stderr().Println("Failed to read file '%s': %s", expandedPath, err.Error())
		return nil, err
	}
	return contents, nil
}

func LoadBytesFromUrl(path string) ([]byte, error) {
	// Get the data
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// isValidUrl tests a string to determine if it is a url or not.
func (l *LocalRegistry) isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}