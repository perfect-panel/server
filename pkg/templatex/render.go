package templatex

import (
	"bytes"
	"text/template"
)

func RenderToString(tmpl string, data map[string]interface{}) (string, error) {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
