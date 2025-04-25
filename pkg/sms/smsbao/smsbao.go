package smsbao

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/perfect-panel/server/pkg/templatex"
	"github.com/perfect-panel/server/pkg/tool"
)

const BaseURL = "https://api.smsbao.com"

type Config struct {
	Access   string `json:"access"`
	Secret   string `json:"secret"`
	Template string `json:"template"`
}

type Client struct {
	config *Config
	client *resty.Client
}

func NewClient(config Config) *Client {
	client := resty.New()
	client.SetBaseURL(BaseURL)
	return &Client{
		config: &config,
		client: client,
	}
}

func (c *Client) SendCode(area, mobile, code string) error {
	apiUrl := "/sms"
	text, err := templatex.RenderToString(c.config.Template, map[string]interface{}{
		"code": code,
	})
	if err != nil {
		return fmt.Errorf("failed to render sms template: %s", err.Error())
	}
	param := map[string]string{
		"u": c.config.Access,
		"p": tool.Md5Encode(c.config.Secret, false),
		"m": mobile,
		"c": text,
	}
	if area != "86" {
		apiUrl = "/wsms"
		param["m"] = fmt.Sprintf("+%s%s", area, mobile)
	}
	resp, err := c.client.R().SetQueryParams(param).Get(apiUrl)
	if err != nil {
		return err
	}
	err = parseError(resp.Body())
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetSendCodeContent(code string) string {
	text, _ := templatex.RenderToString(c.config.Template, map[string]interface{}{
		"code": code,
	})
	return text
}
