package vproxy

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/perfect-panel/server/pkg/adapter/general"
	"github.com/perfect-panel/server/pkg/adapter/proxy"
	"github.com/perfect-panel/server/pkg/traffic"
)

type UserInfo struct {
	Upload       int64
	Download     int64
	TotalTraffic int64
	ExpiredDate  time.Time
}

func BuildVproxy(servers []proxy.Proxy, uuid string, userinfo UserInfo) []byte {
	upload := traffic.AutoConvert(userinfo.Upload, false)
	download := traffic.AutoConvert(userinfo.Download, false)
	total := traffic.AutoConvert(userinfo.TotalTraffic, false)
	expiredAt := userinfo.ExpiredDate.Format("2006-01-02 15:04:05")
	uri := fmt.Sprintf("上行:%s，下行:%s，总共:%s，到期:%s\r\n", upload, download, total, expiredAt)
	for _, s := range servers {
		switch s.Protocol {
		case "vmess":
			uri += general.VmessUri(s, uuid) + "\r\n"
		case "shadowsocks":
			uri += general.ShadowsocksUri(s, uuid) + "\r\n"
		case "trojan":
			uri += general.TrojanUri(s, uuid) + "\r\n"
		case "vless":
			uri += general.VlessUri(s, uuid) + "\r\n"
		case "hysteria2":
			uri += general.Hysteria2Uri(s, uuid) + "\r\n"
		// case "tuic":
		// 	uri += general.TuicUri(s, uuid) + "\r\n"
		default:
			continue
		}
	}

	return []byte(base64.StdEncoding.EncodeToString([]byte(uri)))
}
