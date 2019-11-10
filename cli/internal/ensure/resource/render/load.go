package render

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/solo-io/go-utils/osutils"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
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




