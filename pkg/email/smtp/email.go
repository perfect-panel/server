package smtp

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

type Client struct {
	conf   Config
	dailer *gomail.Dialer
}
type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	From     string `json:"from"`
	SSL      bool   `json:"ssl"`
	SiteName string `json:"siteName"`
}

func NewClient(conf *Config) *Client {
	if conf == nil {
		return nil
	}
	dailer := gomail.NewDialer(conf.Host, conf.Port, conf.User, conf.Pass)
	dailer.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		ServerName:         conf.Host,
	}

	return &Client{conf: *conf, dailer: dailer}
}

func (m *Client) Send(to []string, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", m.conf.From, m.conf.SiteName)
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	return m.dailer.DialAndSend(msg)
}
