package twilio

import "testing"

func TestClient_SendCode(t *testing.T) {
	t.Skipf("Skip TestClient_SendCode test")
	client := NewClient(Config{
		Access: "", Secret: "", PhoneNumber: "", Template: "",
	})
	err := client.SendCode("", "", "123456")
	if err != nil {
		t.Errorf("SendCode() error = %v", err.Error())
		return
	}
	t.Logf("SendCode() success")
}
