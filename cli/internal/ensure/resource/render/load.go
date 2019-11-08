package render

import (
	"bytes"
	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

func LoadTemplate(tmpl string, values Values) (string, error) {
	parsed, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}
	out := strings.Builder{}
	vals, err := renderValues(values)
	if err != nil {
		return "", err
	}
	err = parsed.Execute(&out, vals)
	return out.String(), err
}

func renderValues(values Values) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	for k := range values {
		v, err := values.GetValue(k)
		if err != nil {
			return nil, err
		}
		vals[k] = v
	}
	return vals, nil
}

func LoadFile(path string) (string, error) {
	b, err := LoadBytes(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func LoadBytes(path string) ([]byte, error) {
	if isValidUrl(path) {
		contents, err := loadBytesFromUrl(path)
		if err == nil {
			return contents, nil
		}
	}

	osClient := osutils.NewOsClient()
	contents, err := osClient.ReadFile(path)
	if err != nil {
		cmd.Stderr().Println("Failed to read file '%s': %s", path, err.Error())
		return nil, err
	}
	return contents, nil
}

func loadBytesFromUrl(path string) ([]byte, error) {
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
func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}