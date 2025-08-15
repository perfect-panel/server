package adapter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/pkg/tool"
)

func adapterProxy(svr server.Server, host string, port uint64) (Proxy, error) {
	tags := strings.Split(svr.Tags, ",")
	if len(tags) > 0 {
		tags = tool.RemoveDuplicateElements(tags...)
	}
	node := Proxy{
		Name: svr.Name,
		Host: host,
		Port: port,
		Type: svr.Protocol,
		Tags: tags,
	}
	switch svr.Protocol {
	case "shadowsocks":
		var ss server.Shadowsocks
		if err := json.Unmarshal([]byte(svr.Config), &ss); err != nil {
			return node, fmt.Errorf("unmarshal shadowsocks config: %v", err.Error())
		}
		if port == 0 {
			node.Port = uint64(ss.Port)
		}
		node.Method = ss.Method
		node.ServerKey = ss.ServerKey
	case "vless":
		var vless server.Vless
		if err := json.Unmarshal([]byte(svr.Config), &vless); err != nil {
			return node, fmt.Errorf("unmarshal vless config: %v", err.Error())
		}
		if port == 0 {
			node.Port = uint64(vless.Port)
		}
		node.Flow = vless.Flow
		node.Transport = vless.Transport
		tool.DeepCopy(&node, vless.TransportConfig)
		node.Security = vless.Security
		tool.DeepCopy(&node, vless.SecurityConfig)
	case "vmess":
		var vmess server.Vmess
		if err := json.Unmarshal([]byte(svr.Config), &vmess); err != nil {
			return node, fmt.Errorf("unmarshal vmess config: %v", err.Error())
		}
		if port == 0 {
			node.Port = uint64(vmess.Port)
		}
		node.Flow = vmess.Flow
		node.Transport = vmess.Transport
		tool.DeepCopy(&node, vmess.TransportConfig)
		node.Security = vmess.Security
		tool.DeepCopy(&node, vmess.SecurityConfig)
	case "trojan":
		var trojan server.Trojan
		if err := json.Unmarshal([]byte(svr.Config), &trojan); err != nil {
			return node, fmt.Errorf("unmarshal trojan config: %v", err.Error())
		}
		if port == 0 {
			node.Port = uint64(trojan.Port)

		}

		node.Flow = trojan.Flow
		node.Transport = trojan.Transport
		tool.DeepCopy(&node, trojan.TransportConfig)
		node.Security = trojan.Security
		tool.DeepCopy(&node, trojan.SecurityConfig)
	case "hysteria2":
		var hysteria2 server.Hysteria2
		if err := json.Unmarshal([]byte(svr.Config), &hysteria2); err != nil {
			return node, fmt.Errorf("unmarshal hysteria2 config: %v", err.Error())
		}
		if port == 0 {
			node.Port = uint64(hysteria2.Port)
		}
		node.HopPorts = hysteria2.HopPorts
		node.HopInterval = hysteria2.HopInterval
		node.ObfsPassword = hysteria2.ObfsPassword
		tool.DeepCopy(&node, hysteria2.SecurityConfig)
	case "tuic":
		var tuic server.Tuic
		if err := json.Unmarshal([]byte(svr.Config), &tuic); err != nil {
			return node, fmt.Errorf("unmarshal tuic config: %v", err.Error())
		}
		if port == 0 {
			node.Port = uint64(tuic.Port)
		}
		node.DisableSNI = tuic.DisableSNI
		node.ReduceRtt = tuic.ReduceRtt
		node.UDPRelayMode = tuic.UDPRelayMode
		node.CongestionController = tuic.CongestionController
	case "anytls":
		var anytls server.AnyTLS
		if err := json.Unmarshal([]byte(svr.Config), &anytls); err != nil {
			return node, fmt.Errorf("unmarshal anytls config: %v", err.Error())
		}
		if port == 0 {
			node.Port = uint64(anytls.Port)
		}
		tool.DeepCopy(&node, anytls.SecurityConfig)
	default:
		return node, fmt.Errorf("unsupported protocol: %s", svr.Protocol)
	}
	return node, nil
}
