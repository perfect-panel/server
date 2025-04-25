package sms

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/perfect-panel/server/pkg/sms/abosend"
	"github.com/perfect-panel/server/pkg/sms/alibabacloud"
	"github.com/perfect-panel/server/pkg/sms/smsbao"
	"github.com/perfect-panel/server/pkg/sms/twilio"
)

type Sender interface {
	SendCode(area, mobile, code string) error
	GetSendCodeContent(code string) string
}

func NewSender(platform, config string) (Sender, error) {
	log.Printf("platform: %s, config: %s", platform, config)
	switch parsePlatform(platform) {
	case AlibabaCloud:
		cfg := alibabacloud.Config{}
		if err := json.Unmarshal([]byte(config), &cfg); err != nil {
			return nil, fmt.Errorf("alibabacloud config unmarshal failed: %v", err.Error())
		}
		return alibabacloud.NewClient(cfg), nil
	case Abosend:
		cfg := abosend.Config{}
		if err := json.Unmarshal([]byte(config), &cfg); err != nil {
			return nil, fmt.Errorf("abosend config unmarshal failed: %v", err.Error())
		}
		return abosend.NewClient(cfg), nil
	case Smsbao:
		cfg := smsbao.Config{}
		if err := json.Unmarshal([]byte(config), &cfg); err != nil {
			return nil, fmt.Errorf("smsbao config unmarshal failed: %v", err.Error())
		}
		return smsbao.NewClient(cfg), nil
	case Twilio:
		cfg := twilio.Config{}
		if err := json.Unmarshal([]byte(config), &cfg); err != nil {
			return nil, fmt.Errorf("twilio config unmarshal failed: %v", err.Error())
		}
		return twilio.NewClient(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}
