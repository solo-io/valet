package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"os"
	"strings"
	"text/template"
)

type Template struct {
	Path      string            `yaml:"path"`
	Values    map[string]string `yaml:"values"`
	EnvValues map[string]string `yaml:"envValues"`
}

type TemplateValue struct {
	Value  string `yaml:"value"`
	EnvVar string `yaml:"envVar"`
}

func (g *Template) setValue(key, value string) {
	if g.Values == nil {
		g.Values = make(map[string]string)
	}
	g.Values[key] = value
}

func (g *Template) Ensure(ctx context.Context, command cmd.Factory) error {
	rendered, err := g.render(ctx)
	if err != nil {
		return err
	}
	return command.Kubectl().ApplyStdIn(rendered).Cmd().Run(ctx)
}

func (g *Template) Teardown(ctx context.Context, command cmd.Factory) error {
	rendered, err := g.render(ctx)
	if err != nil {
		return err
	}
	return command.Kubectl().DeleteStdIn(rendered).Cmd().Run(ctx)
}

func (g *Template) render(ctx context.Context) (string, error) {
	tmpl, err := LoadFile(g.Path)
	if err != nil {
		return "", err
	}
	parsed, err := template.New(g.Path).Parse(tmpl)
	if err != nil {
		return "", err
	}
	out := strings.Builder{}
	values, err := g.renderValues(ctx)
	if err != nil {
		return "", err
	}
	err = parsed.Execute(&out, values)
	return out.String(), err
}

func (g *Template) renderValues(ctx context.Context) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	for k, v := range g.Values {
		values[k] = v
	}
	for k, v := range g.EnvValues {
		values[k] = os.Getenv(v)
	}
	return values, nil
}

func LoadFile(path string) (string, error) {
	b, err := loadBytesFromPath(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
