package subscribe

// V4.3 决策 25 + V4.4 §41 — 11 客户端 UA → 模板选择 + adapter 输出。
// 不依赖 DB / Redis,跑 in-process 验证。
//
// 测试目标:
//   1. 11 个真实 client UA(取自各 client app 公开版本)→ PickClientApp 选对模板
//   2. UA 大小写不敏感
//   3. Stash 不会被 Clash 抢(决策 25 边界)
//   4. 未知 UA → 走 IsDefault
//   5. 跑通 adapter Build 三种 output 格式

import (
	"strings"
	"testing"
	"time"

	"github.com/perfect-panel/server/adapter"
	"github.com/perfect-panel/server/internal/model/client"
	"github.com/perfect-panel/server/internal/model/node"
)

// 11 款客户端 + 默认 — 与 V4.3 §13 FAQ 列表对齐。
// UserAgent 字段是模板 UA 子串,匹配 lowercase contains。
//
// 顺序敏感(specificity):更长 / 更具体的 UA 子串排前面,
//   - Stash 必须在 Clash 前(Stash UA 含 "clash")
//   - FlClash 必须在 Clash 前("flclash" 含 "clash")
//   - ClashMeta 必须在 Clash 前("clashmeta" 含 "clash")
//   - Surge Mac 必须在 Surge 前("surge mac" 含 "surge")
func sampleApps() []*client.SubscribeApplication {
	return []*client.SubscribeApplication{
		{Id: 1, Name: "v2rayN", UserAgent: "v2rayn", OutputFormat: "base64", SubscribeTemplate: tplBase64()},
		{Id: 2, Name: "Stash", UserAgent: "stash", OutputFormat: "yaml", SubscribeTemplate: tplYaml()},
		{Id: 3, Name: "Clash Meta for Android", UserAgent: "clashmeta", OutputFormat: "yaml", SubscribeTemplate: tplYaml()},
		{Id: 9, Name: "FlClash", UserAgent: "flclash", OutputFormat: "yaml", SubscribeTemplate: tplYaml()},
		{Id: 4, Name: "Clash", UserAgent: "clash", OutputFormat: "yaml", SubscribeTemplate: tplYaml()},
		{Id: 5, Name: "Shadowrocket", UserAgent: "shadowrocket", OutputFormat: "base64", SubscribeTemplate: tplBase64()},
		{Id: 6, Name: "Hiddify", UserAgent: "hiddify", OutputFormat: "json", SubscribeTemplate: tplJson()},
		{Id: 7, Name: "Quantumult X", UserAgent: "quantumult", OutputFormat: "conf", SubscribeTemplate: tplConf()},
		{Id: 8, Name: "Loon", UserAgent: "loon", OutputFormat: "conf", SubscribeTemplate: tplConf()},
		{Id: 10, Name: "Surge for Mac", UserAgent: "surge mac", OutputFormat: "conf", SubscribeTemplate: tplConf()},
		{Id: 11, Name: "Surge", UserAgent: "surge", OutputFormat: "conf", SubscribeTemplate: tplConf()},
		{Id: 99, Name: "Default", UserAgent: "ppanel-default", OutputFormat: "base64", SubscribeTemplate: tplBase64(), IsDefault: true},
	}
}

func tplYaml() string {
	return "proxies:\n{{ range .Proxies }}  - name: {{ .Name }}\n{{ end }}\nrules:\n{{ range .DirectList }}  - DOMAIN-SUFFIX,{{ . }},DIRECT\n{{ end }}  - MATCH,PROXY\n"
}
func tplBase64() string {
	return "{{ range .Proxies }}vless://{{ $.UserInfo.Password }}@{{ .Server }}:{{ .Port }}#{{ .Name }}\n{{ end }}"
}
func tplJson() string {
	return `{"site":"{{ .SiteName }}","proxies":[{{ range $i, $p := .Proxies }}{{ if $i }},{{ end }}{"n":"{{ $p.Name }}"}{{ end }}]}`
}
func tplConf() string {
	return "[Proxy]\n{{ range .Proxies }}{{ .Name }} = vless, {{ .Server }}, {{ .Port }}\n{{ end }}"
}

func sampleNodes() []*node.Node {
	// adapter.Proxies 需 server.Protocols 是 JSON array,protocol.Type 与 node.Protocol 匹配。
	srv := &node.Server{
		Id:         1,
		Address:    "1.2.3.4",
		DirectList: `["stripe.com","alipay.com"]`,
		Protocols:  `[{"type":"vless","security":"tls","sni":"jp1.example.com"}]`,
	}
	return []*node.Node{
		{Id: 1, Name: "JP-1", ServerId: 1, Address: "jp1.example.com", Port: 443, Protocol: "vless", Server: srv},
	}
}

// ---- 测试主体 ----

func TestPickClientApp_AllElevenClients(t *testing.T) {
	apps := sampleApps()
	cases := []struct {
		ua       string
		expectId int64
	}{
		// 真实 UA 串(精简到关键子串 + 版本号)
		{"v2rayN/6.45 (Windows 10)", 1},
		{"Stash/2.0.0", 2},
		{"ClashMeta-Android/2.10.4 ClashMeta/v1.18.0", 3},
		{"ClashforWindows/0.20.39", 4},
		{"Shadowrocket/1.x ConnectionDelegate", 5},
		{"Hiddify-Android/0.14.5 (com.hiddify.android)", 6},
		{"Quantumult%20X/1.0.30 CFNetwork", 7},
		{"Loon/766 CFNetwork/1492.0.1 Darwin/23.3.0", 8},
		{"FlClash/0.8.50", 9},
		{"Surge Mac/2543 (com.nssurge.surge-mac)", 10},
		{"Surge iOS/2543", 11},
	}
	for _, tc := range cases {
		t.Run(tc.ua, func(t *testing.T) {
			got := PickClientApp(tc.ua, apps)
			if got == nil {
				t.Fatalf("UA %q got nil app", tc.ua)
			}
			if got.Id != tc.expectId {
				t.Errorf("UA %q: expected app id %d, got %d (%s)", tc.ua, tc.expectId, got.Id, got.Name)
			}
		})
	}
}

func TestPickClientApp_StashDoesNotMatchClash(t *testing.T) {
	apps := sampleApps()
	// Stash UA 自带 "clash" 子串(老版本 Stash UA: "Stash/0.41.0 Clash/v1.18")
	got := PickClientApp("Stash/0.41.0 Clash/v1.18", apps)
	if got == nil || got.Name != "Stash" {
		t.Fatalf("Stash with embedded 'Clash' should match Stash, got %+v", got)
	}
}

func TestPickClientApp_UnknownFallsBackToDefault(t *testing.T) {
	apps := sampleApps()
	got := PickClientApp("MozillaCurl/9.9.9 No-such-client", apps)
	if got == nil || !got.IsDefault {
		t.Fatalf("unknown UA should fall back to default, got %+v", got)
	}
}

func TestPickClientApp_EmptyClientsReturnsNil(t *testing.T) {
	if got := PickClientApp("Clash", nil); got != nil {
		t.Errorf("expected nil for empty client list, got %+v", got)
	}
}

func TestPickClientApp_CaseInsensitive(t *testing.T) {
	apps := sampleApps()
	// 全大写 UA 也能匹配
	got := PickClientApp("CLASH/1.0", apps)
	if got == nil || got.Name != "Clash" {
		t.Fatalf("uppercase UA must still match Clash, got %+v", got)
	}
}

// ---- adapter 端到端:每种 output_format 都跑一遍 Build,确保模板渲染不炸 ----

func TestAdapter_Build_AllFormats(t *testing.T) {
	apps := sampleApps()
	servers := sampleNodes()
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	user := adapter.User{
		Password:     "00000000-0000-0000-0000-000000000000",
		ExpiredAt:    now.AddDate(0, 0, 30),
		Download:     1024,
		Upload:       1024,
		Traffic:      107374182400,
		SubscribeURL: "https://sub.example.com/sub?token=abc",
	}
	for _, app := range apps {
		t.Run(app.Name+"/"+app.OutputFormat, func(t *testing.T) {
			a := adapter.NewAdapter(
				app.SubscribeTemplate,
				adapter.WithServers(servers),
				adapter.WithSiteName("PerfectPanel"),
				adapter.WithSubscribeName("Premium"),
				adapter.WithOutputFormat(app.OutputFormat),
				adapter.WithUserInfo(user),
				adapter.WithDirectList([]string{"stripe.com", "alipay.com"}),
			)
			c, err := a.Client()
			if err != nil {
				t.Fatalf("Client(): %v", err)
			}
			out, err := c.Build()
			if err != nil {
				t.Fatalf("Build(): %v", err)
			}
			if len(out) == 0 {
				t.Fatal("empty build output")
			}
			// 输出语义抽查:
			s := string(out)
			switch app.OutputFormat {
			case "yaml":
				// V4.3 决策 39:DirectList 出在 yaml 模板的 rules 里
				if !strings.Contains(s, "DOMAIN-SUFFIX,stripe.com,DIRECT") {
					t.Errorf("yaml output missing direct rule: %q", s)
				}
			case "json":
				if !strings.Contains(s, `"site":"PerfectPanel"`) {
					t.Errorf("json output missing SiteName: %q", s)
				}
			case "conf":
				if !strings.Contains(s, "[Proxy]") {
					t.Errorf("conf output missing [Proxy]: %q", s)
				}
			case "base64":
				// base64 模板下 Build 会做 base64 包装
				if strings.Contains(s, "vless://") {
					t.Errorf("base64 mode should be encoded, found raw vless: %q", s)
				}
			}
		})
	}
}

// ---- 直连白名单注入:不同模板都能渲染 ----

func TestAdapter_DirectListEmptyOK(t *testing.T) {
	a := adapter.NewAdapter(
		tplYaml(),
		adapter.WithServers(sampleNodes()),
		adapter.WithOutputFormat("yaml"),
		adapter.WithUserInfo(adapter.User{Password: "x"}),
		adapter.WithDirectList(nil),
	)
	c, _ := a.Client()
	out, err := c.Build()
	if err != nil {
		t.Fatalf("empty direct list should not break Build: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "MATCH,PROXY") {
		t.Errorf("yaml output missing fallback rule: %q", s)
	}
	if strings.Contains(s, "DOMAIN-SUFFIX") {
		t.Errorf("empty DirectList should not emit any DOMAIN-SUFFIX rule: %q", s)
	}
}
