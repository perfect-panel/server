package adapter

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/pkg/adapter/surfboard"
)

func createTestServer() []*server.Server {
	c := server.Shadowsocks{
		Method:    "aes-256-gcm",
		Port:      10301,
		ServerKey: "",
	}
	data, _ := json.Marshal(c)

	relays := creatRelayNode()
	relay, _ := json.Marshal(relays)
	enable := true
	// 创建一个测试用的服务器列表
	return []*server.Server{
		{
			Id:           1,
			Name:         "Test Server 1",
			Tags:         "",
			Country:      "CN",
			City:         "",
			Latitude:     "",
			Longitude:    "",
			ServerAddr:   "test1.example.com",
			RelayMode:    "random",
			RelayNode:    string(relay),
			SpeedLimit:   0,
			TrafficRatio: 0,
			GroupId:      0,
			Protocol:     "shadowsocks",
			Config:       string(data),
			Enable:       &enable,
			Sort:         0,
		},
	}
}
func creatRelayNode() []*server.NodeRelay {
	var nodes []*server.NodeRelay
	for i := 0; i < 10; i++ {
		port := 10301 + i
		c := server.NodeRelay{
			Host:   fmt.Sprintf("192.168.1.%d", i),
			Port:   port,
			Prefix: fmt.Sprintf("relay-%d", i),
		}
		nodes = append(nodes, &c)
	}
	return nodes
}

func TestNewAdapter(t *testing.T) {
	nodes := createTestServer()

	rules := []*server.RuleGroup{
		{
			Name:  "Test Rule Group 1",
			Tags:  "",
			Rules: "DOMAIN-SUFFIX,example.com,Test Rule Group 1",
		},
	}

	adapter := NewAdapter(nodes, rules)
	bytes, err := adapter.BuildClash("some-uuid")
	if err != nil {
		t.Errorf("Failed to build adapter: %v", err)
		return
	}
	t.Logf("Adapter built successfully: %s", string(bytes))
}

func TestAdapter_BuildSingbox(t *testing.T) {
	nodes := createTestServer()

	rules := []*server.RuleGroup{
		{
			Name:  "Test Rule Group 1",
			Tags:  "",
			Rules: "DOMAIN-SUFFIX,example.com,Test Rule Group 1",
		},
	}

	adapter := NewAdapter(nodes, rules)
	bytes, err := adapter.BuildSingbox("some-uuid")
	if err != nil {
		t.Errorf("Failed to build adapter: %v", err)
		return
	}
	var pretty map[string]interface{}
	_ = json.Unmarshal(bytes, &pretty)

	if pretty == nil {
		t.Errorf("Failed to parse Singbox config")
		return
	}

	prettyStr, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		t.Errorf("Failed to format Singbox config: %v", err)
		return
	}
	t.Logf("Adapter built successfully: \n %s", string(prettyStr))
}

func TestAdapter_BuildSurfboard(t *testing.T) {
	nodes := createTestServer()
	rules := []*server.RuleGroup{
		{
			Name:  "Test Rule Group 1",
			Tags:  "",
			Rules: "DOMAIN-SUFFIX,example.com,Test Rule Group 1",
		},
	}
	adapter := NewAdapter(nodes, rules)
	user := surfboard.UserInfo{
		UUID:         "some-uuid",
		Upload:       200,
		Download:     13012,
		TotalTraffic: 1024000,
		ExpiredDate:  time.Now().Add(24 * time.Hour),
		SubscribeURL: "",
	}
	bytes := adapter.BuildSurfboard("test-site", user)
	if bytes == nil {
		t.Errorf("Failed to build adapter")
		return
	}
	t.Logf("Adapter built successfully: %s", string(bytes))
}
