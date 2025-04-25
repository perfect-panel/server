package quantumultx

import (
	"testing"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

func createVMess() proxy.Proxy {

	return proxy.Proxy{
		Name:     "Vmess",
		Server:   "test.xxxx.com",
		Port:     13002,
		Protocol: "vmess",
		Option: proxy.Vmess{
			Port:      13002,
			Transport: "websocket",
			TransportConfig: proxy.TransportConfig{
				Path: "/ws",
				Host: "test.xx.com",
			},
			Security: "none",
		},
	}
}

func createSS() proxy.Proxy {
	return proxy.Proxy{
		Name:     "Shadowsocks",
		Server:   "test.xxxx.com",
		Port:     10301,
		Protocol: "shadowsocks",
		Option: proxy.Shadowsocks{
			Port:      10301,
			Method:    "aes-256-gcm",
			ServerKey: "123456",
		},
	}
}

func createTrojan() proxy.Proxy {

	return proxy.Proxy{
		Name:     "Trojan",
		Server:   "test.xxxx.com",
		Port:     13002,
		Protocol: "trojan",
		Option: proxy.Trojan{
			Port:      13002,
			Transport: "websocket",
			TransportConfig: proxy.TransportConfig{
				Path: "/ws",
				Host: "baidu.com",
			},
			SecurityConfig: proxy.SecurityConfig{
				SNI:           "baidu.com",
				AllowInsecure: true,
			},
		},
	}
}
func TestVmess(t *testing.T) {
	s := createVMess()
	vmess := buildVmess(s, "uuid")
	t.Log(vmess)
	// output:
	// vmess=127.0.0.1:13002,method=chacha20-poly1305,password=uuid,fast-open=true,udp-relay=true,tag=Vmess,tls-verification=true,obfs-uri=/ws,obfs-host=baidu.com
}

func TestShadowsocks(t *testing.T) {
	s := createSS()
	shadowsocks := buildShadowsocks(s, "uuid")
	t.Log(shadowsocks)
	// output:
	// shadowsocks=127.0.0.1:10301,method=aes-256-gcm,password=uuid,fast-open=true,udp-relay=true,tag=Shadowsocks
}

func TestTrojan(t *testing.T) {
	s := createTrojan()
	trojan := buildTrojan(s, "password")
	t.Log(trojan)
	// output:
	// trojan=192.168.0.1:13002,password=password,fast-open=true,udp-relay=true,tag=Trojan,obfs=wss,obfs-uri=ws,obfs-host=baidu.com
}

func TestBuildQuantumultX(t *testing.T) {
	var servers []proxy.Proxy
	uri := BuildQuantumultX(servers, "uuid")
	t.Log(uri)

	// output:
	// c2hhZG93c29ja3M9MTI3LjAuMC4xOjEwMzAxLG1ldGhvZD1hZXMtMjU2LWdjbSxwYXNzd29yZD11dWlkLGZhc3Qtb3Blbj10cnVlLHVkcC1yZWxheT10cnVlLHRhZz1TaGFkb3dzb2Nrcw0KdHJvamFuPTE5Mi4xNjguMC4xOjEzMDAyLHBhc3N3b3JkPXV1aWQsZmFzdC1vcGVuPXRydWUsdWRwLXJlbGF5PXRydWUsdGFnPVRyb2phbixvYmZzPXdzcyxvYmZzLXVyaT13cyxvYmZzLWhvc3Q9YmFpZHUuY29tDQp2bWVzcz0xMjcuMC4wLjE6MTMwMDIsbWV0aG9kPWNoYWNoYTIwLXBvbHkxMzA1LHBhc3N3b3JkPXV1aWQsZmFzdC1vcGVuPXRydWUsdWRwLXJlbGF5PXRydWUsdGFnPVZtZXNzLHRscy12ZXJpZmljYXRpb249dHJ1ZSxvYmZzLXVyaT0vd3Msb2Jmcy1ob3N0PWJhaWR1LmNvbQ0K
}
