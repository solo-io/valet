package render

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/pkg/cmd"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//go:generate mockgen -destination ./mocks/file_store_mock.go github.com/solo-io/valet/pkg/render FileStore

type FileStore interface {
	Load(path string) (string, error)
	LoadYaml(path string, i interface{}) error
	Save(path, contents string) error
}

func NewFileStore() *fileStore {
	return &fileStore{}
}

var _ FileStore = new(fileStore)

type fileStore struct {}

func (f *fileStore) Save(path, contents string) error {
	return ioutil.WriteFile(path, []byte(contents), os.ModePerm)
}

func (f *fileStore) Load(path string) (string, error) {
	return f.loadFile(path)
}

func (f *fileStore) LoadYaml(path string, deserialized interface{}) error {
	bytes, err := f.loadBytes(path)
	if err != nil {
		return err
	}
	return yaml.UnmarshalStrict(bytes, deserialized, yaml.DisallowUnknownFields)
}

func (f *fileStore) loadFile(path string) (string, error) {
	b, err := f.loadBytes(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (f *fileStore) loadBytes(path string) ([]byte, error) {
	if f.isValidUrl(path) {
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
func (f *fileStore) isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}

