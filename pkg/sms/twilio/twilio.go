package twilio

import (
	"fmt"

	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/templatex"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Config struct {
	Access      string `json:"access"`
	Secret      string `json:"secret"`
	PhoneNumber string `json:"phone_number"`
	Template    string `json:"template"`
}

type Client struct {
	config Config
	client *twilio.RestClient
}

func NewClient(config Config) *Client {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: config.Access,
		Password: config.Secret,
	})
	return &Client{
		config: config,
		client: client,
	}
}

func (c *Client) SendCode(area, mobile, code string) error {
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(fmt.Sprintf("+%s%s", area, mobile))
	params.SetFrom(c.config.PhoneNumber)
	text, err := templatex.RenderToString(c.config.Template, map[string]interface{}{
		"code": code,
	})
	if err != nil {
		logger.Error("twilio send code render template error", logger.Field("error", err.Error()), logger.Field("template", c.config.Template), logger.Field("code", code))
	}
	params.SetBody(text)
	resp, err := c.client.Api.CreateMessage(params)
	if err != nil {
		logger.Error("twilio send code error", logger.Field("error", err.Error()), logger.Field("params", params))
		return fmt.Errorf("twilio send code error: %s", err.Error())
	}
	if resp.ErrorCode != nil {
		logger.Error("twilio send code error", logger.Field("error_code", *resp.ErrorCode), logger.Field("error_message", *resp.ErrorMessage))
		return fmt.Errorf("twilio send code error: %s", *resp.ErrorMessage)
	}
	return nil
}

func (c *Client) GetSendCodeContent(code string) string {
	text, _ := templatex.RenderToString(c.config.Template, map[string]interface{}{
		"code": code,
	})
	return text
}
