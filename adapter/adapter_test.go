package adapter

import (
	"testing"
	"time"
)

func TestAdapter_Client(t *testing.T) {
	servers := getServers()
	if len(servers) == 0 {
		t.Errorf("[Test] No servers found")
		return
	}
	a := NewAdapter(tpl, WithServers(servers), WithUserInfo(User{
		Password:     "test-password",
		ExpiredAt:    time.Now().AddDate(1, 0, 0),
		Download:     0,
		Upload:       0,
		Traffic:      1000,
		SubscribeURL: "https://example.com/subscribe",
	}))
	client, err := a.Client()
	if err != nil {
		t.Errorf("[Test] Failed to get client: %v", err.Error())
		return
	}
	bytes, err := client.Build()
	if err != nil {
		t.Errorf("[Test] Failed to build client config: %v", err.Error())
		return
	}
	t.Logf("[Test] Client config built successfully: %s", string(bytes))

}
