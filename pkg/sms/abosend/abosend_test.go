package abosend

import "testing"

func TestNewClient(t *testing.T) {
	t.Skipf("Skip TestNewClient test")
	client := createClient()
	err := client.SendCode("1", "", "223322")
	if err != nil {
		t.Errorf("TestNewClient() error = %v", err.Error())
		return
	}
	t.Logf("TestNewClient success")
}

func TestClient_GetSendCodeContent(t *testing.T) {
	t.Skipf("Skip TestClient_GetSendCodeContent test")
	client := createClient()
	content := client.GetSendCodeContent("223322")
	t.Logf("TestClient_GetSendCodeContent() = %v", content)
}

func createClient() *Client {
	return NewClient(Config{
		ApiDomain: "https://smsapi.abosend.com",
		Access:    "",
		Secret:    "",
		Template:  "您的验证码是：{{.code}}。请不要把验证码泄露给其他人。",
	})
}
