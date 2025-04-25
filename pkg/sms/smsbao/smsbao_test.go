package smsbao

import "testing"

func TestNewClient(t *testing.T) {
	t.Skipf("Skip TestNewClient test")
	client := NewClient(Config{
		Template: "【XXX】您的验证码是：{{.code}}，有效期 {{.expiration}}。请不要把验证码泄露给其他人。",
	})
	err := client.SendCode("1", "", "223322")
	if err != nil {
		t.Errorf("TestNewClient() error = %v", err.Error())
		return
	}
	t.Logf("TestNewClient success")
}
