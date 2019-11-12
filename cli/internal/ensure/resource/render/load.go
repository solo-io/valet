package render

import (
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"strings"
	"text/template"
)

func LoadTemplate(tmpl string, values Values, runner cmd.Runner) (string, error) {
	parsed, err := template.New("").Option("missingkey=error").Parse(tmpl)
	if err != nil {
		return "", err
	}
	out := strings.Builder{}
	vals, err := values.Render(runner)
	if err != nil {
		return "", err
	}
	err = parsed.Execute(&out, vals)
	return out.String(), err
}




