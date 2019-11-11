package render

import (
	"strings"
	"text/template"
)

func LoadTemplate(tmpl string, values Values) (string, error) {
	parsed, err := template.New("").Option("missingkey=error").Parse(tmpl)
	if err != nil {
		return "", err
	}
	out := strings.Builder{}
	vals, err := values.Render()
	if err != nil {
		return "", err
	}
	err = parsed.Execute(&out, vals)
	return out.String(), err
}




