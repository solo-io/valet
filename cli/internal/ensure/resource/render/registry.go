package render

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

//go:generate mockgen -destination ./mocks/registry_mock.go github.com/solo-io/valet/cli/internal/ensure/resource/render Registry

const (
	DefaultRegistry = "default"
)

var (
	UnknownRegistryError = func(name string) error {
		return errors.Errorf("Unknown registry %s", name)
	}
)

type Registry interface {
	LoadFile(path string) (string, error)
}

type LocalRegistry struct {
	WorkingDirectory string `yaml:"dir"`
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
	expandedPath := expandEnv(path)
	contents, err := osClient.ReadFile(expandedPath)
	if err != nil {
		cmd.Stderr().Println("Failed to read file '%s': %s", expandedPath, err.Error())
		return nil, err
	}
	return contents, nil
}

func expandEnv(path string) string {
	if strings.HasPrefix(path, "~") {
		path = strings.TrimPrefix(path, "~")
		path = fmt.Sprintf("$HOME%s", path)
	}
	return os.ExpandEnv(path)
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
