package email

import (
	"bytes"
	"html/template"
	"testing"
)

type VerifyTemplate struct {
	Type     uint8
	SiteLogo string
	SiteName string
	Expire   uint8
	Code     string
}

func TestVerifyEmail(t *testing.T) {
	t.Skipf("Skip TestVerifyEmail test")
	data := VerifyTemplate{
		Type:     1,
		SiteLogo: "https://www.google.com",
		SiteName: "Google",
		Expire:   5,
		Code:     "123456",
	}
	tpl, err := template.New("email").Parse(DefaultEmailVerifyTemplate)
	if err != nil {
		t.Error(err)
	}
	var result bytes.Buffer
	err = tpl.Execute(&result, data)
	if err != nil {
		t.Error(err)
	}
	t.Log(result.String())
}
