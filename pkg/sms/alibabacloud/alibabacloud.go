package alibabacloud

import (
	"encoding/json"
	"fmt"

	"github.com/perfect-panel/server/pkg/logger"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

type Config struct {
	Access       string `json:"access"`
	Secret       string `json:"secret"`
	SignName     string `json:"sign_name"`
	Endpoint     string `json:"endpoint"`
	TemplateCode string `json:"template_code"`
}

type Client struct {
	config *Config
	client *dysmsapi.Client
}

func NewClient(config Config) *Client {
	client, err := initApiClient(config)
	if err != nil {
		logger.Error("NewClient: init Alibaba Cloud Api Client failed", logger.Field("error", err.Error()))
	}
	return &Client{
		config: &config,
		client: client,
	}
}

func (c *Client) SendCode(area, mobile, code string) error {
	if c.client == nil {
		return fmt.Errorf("alibaba cloud api client is nil")
	}
	jsonCode, _ := json.Marshal(map[string]interface{}{
		"code": code,
	})
	sendSmsRequest := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(fmt.Sprintf("%s%s", area, mobile)),
		TemplateCode:  tea.String(c.config.TemplateCode),
		SignName:      tea.String(c.config.SignName),
		TemplateParam: tea.String(string(jsonCode)),
	}
	sendSmsResponse, err := c.client.SendSms(sendSmsRequest)
	if err != nil {
		return err
	}
	if *sendSmsResponse.Body.Code != "OK" {
		return fmt.Errorf("alibaba cloud send sms failed, code: %s, message: %s", *sendSmsResponse.Body.Code, *sendSmsResponse.Body.Message)
	}
	return nil
}

func initApiClient(config Config) (*dysmsapi.Client, error) {
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(config.Access),
		AccessKeySecret: tea.String(config.Secret),
		Endpoint:        tea.String(config.Endpoint),
	}
	if config.Endpoint == "" {
		cfg.Endpoint = tea.String("dysmsapi.ap-southeast-1.aliyuncs.com")
	}
	result, err := dysmsapi.NewClient(cfg)
	return result, err
}

func (c *Client) GetSendCodeContent(code string) string {
	return fmt.Sprintf("TemplateId: %s, SignName:%s, Code: %s", c.config.TemplateCode, c.config.SignName, code)
}
