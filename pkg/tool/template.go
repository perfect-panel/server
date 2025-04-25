package tool

import (
	"bytes"
	"text/template"
)

func RenderTemplateToString(tmpl string, data interface{}) (string, error) {
	// 解析模板
	t, err := template.New("template").Parse(tmpl)
	if err != nil {
		return "", err
	}

	// 创建缓冲区存储结果
	var buf bytes.Buffer

	// 执行模板
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	// 返回结果字符串
	return buf.String(), nil
}
