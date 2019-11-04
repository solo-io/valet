package resource

import (
	"context"
	"os"
	"strings"
	"text/template"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Template struct {
	Path      string            `yaml:"path"`
	Values    map[string]string `yaml:"values"`
	EnvValues map[string]string `yaml:"envValues"`
}

func (t *Template) setValue(key, value string) {
	if t.Values == nil {
		t.Values = make(map[string]string)
	}
	t.Values[key] = value
}

func (t *Template) setEnvValue(key, value string) {
	if t.EnvValues == nil {
		t.EnvValues = make(map[string]string)
	}
	t.EnvValues[key] = value
}

func (t *Template) Ensure(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Ensuring template %s", t.Path)
	rendered, err := t.Load()
	if err != nil {
		return err
	}
	return command.Kubectl().ApplyStdIn(rendered).Cmd().Run(ctx)
}

func (t *Template) Teardown(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Tearing down template %s", t.Path)
	rendered, err := t.Load()
	if err != nil {
		return err
	}
	return command.Kubectl().DeleteStdIn(rendered).Cmd().Run(ctx)
}

func (t *Template) Load() (string, error) {
	tmpl, err := LoadFile(t.Path)
	if err != nil {
		return "", err
	}
	return LoadTemplate(tmpl, t.Values, t.EnvValues)
}

func LoadTemplate(tmpl string, values, envValues map[string]string) (string, error) {
	parsed, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}
	out := strings.Builder{}
	vals := renderValues(values, envValues)
	err = parsed.Execute(&out, vals)
	return out.String(), err
}

func renderValues(values, envValues map[string]string) map[string]interface{} {
	vals := make(map[string]interface{})
	for k, v := range values {
		vals[k] = v
	}
	for k, v := range envValues {
		vals[k] = os.Getenv(v)
	}
	return vals
}

func LoadFile(path string) (string, error) {
	b, err := loadBytesFromPath(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
