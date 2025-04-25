package abosend

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/perfect-panel/server/pkg/random"
	"github.com/perfect-panel/server/pkg/templatex"
	"github.com/perfect-panel/server/pkg/tool"
)

const BaseURL = "https://smsapi.abosend.com"

type Config struct {
	ApiDomain string `json:"api_domain"`
	Access    string `json:"access"`
	Secret    string `json:"secret"`
	Template  string `json:"template"`
}

type Client struct {
	config *Config
	client *resty.Client
}

type request struct {
	OrgCode    string `json:"orgCode"`
	MobileArea string `json:"mobileArea"`
	Mobile     string `json:"mobiles"`
	Content    string `json:"content"`
	Rand       string `json:"rand"`
	Sign       string `json:"sign"`
}

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		SendCode string `json:"sendCode"`
	}
}

func (l *response) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &l)
}

func NewClient(config Config) *Client {
	client := resty.New()
	if config.ApiDomain != "" {
		client.SetBaseURL(config.ApiDomain)
	} else {
		client.SetBaseURL(BaseURL)
	}
	return &Client{
		config: &config,
		client: client,
	}
}

func (c *Client) SendCode(area, mobile, code string) error {
	text, err := templatex.RenderToString(c.config.Template, map[string]interface{}{
		"code": code,
	})
	if err != nil {
		return fmt.Errorf("failed to render sms template: %s", err.Error())
	}
	randNumber := random.Key(6, 0)
	sign := tool.Md5Encode(fmt.Sprintf("%s%s%s%s", c.config.Access, text, randNumber, c.config.Secret), true)
	req := request{
		OrgCode:    c.config.Access,
		MobileArea: fmt.Sprintf("+%s", area),
		Mobile:     fmt.Sprintf("%s%s", area, mobile),
		Content:    text,
		Rand:       randNumber,
		Sign:       sign,
	}
	resp, err := c.client.R().SetBody(req).ForceContentType("application/json").Post("/v2/api/sendSMS")
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("send sms failed, status code: %d", resp.StatusCode())
	}
	var result response
	err = result.Unmarshal(resp.Body())
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %s", err.Error())
	}
	if result.Code != 200 {
		return fmt.Errorf("send sms failed, code: %d, msg: %s", result.Code, result.Message)
	}
	return nil
}

func (c *Client) GetSendCodeContent(code string) string {
	text, _ := templatex.RenderToString(c.config.Template, map[string]interface{}{
		"code": code,
	})
	return text
}
