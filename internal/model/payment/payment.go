package payment

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type Payment struct {
	Id          int64  `gorm:"primaryKey"`
	Name        string `gorm:"type:varchar(100);not null;default:'';comment:Payment Name"`
	Platform    string `gorm:"<-:create;type:varchar(100);not null;comment:Payment Platform"`
	Icon        string `gorm:"type:varchar(255);default:'';comment:Payment Icon"`
	Domain      string `gorm:"type:varchar(255);default:'';comment:Notification Domain"`
	Config      string `gorm:"type:text;not null;comment:Payment Configuration"`
	Description string `gorm:"type:text;comment:Payment Description"`
	FeeMode     uint   `gorm:"type:tinyint(1);not null;default:0;comment:Fee Mode: 0: No Fee 1: Percentage 2: Fixed Amount 3: Percentage + Fixed Amount"`
	FeePercent  int64  `gorm:"type:int;default:0;comment:Fee Percentage"`
	FeeAmount   int64  `gorm:"type:int;default:0;comment:Fixed Fee Amount"`
	Enable      *bool  `gorm:"type:tinyint(1);not null;default:0;comment:Is Enabled"`
	Token       string `gorm:"type:varchar(255);unique;not null;default:'';comment:Payment Token"`
}

func (*Payment) TableName() string {
	return "payment"
}

func (l *Payment) BeforeDelete(_ *gorm.DB) (err error) {
	if l.Id == -1 {
		return fmt.Errorf("can't delete default payment method")
	}
	return nil
}

type Filter struct {
	Mark   string
	Enable *bool
	Search string
}

type StripeConfig struct {
	PublicKey     string `json:"public_key"`
	SecretKey     string `json:"secret_key"`
	WebhookSecret string `json:"webhook_secret"`
	Payment       string `json:"payment"`
}

func (l *StripeConfig) Marshal() ([]byte, error) {
	type Alias StripeConfig
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	})
}

func (l *StripeConfig) Unmarshal(data []byte) error {
	type Alias StripeConfig
	aux := (*Alias)(l)
	return json.Unmarshal(data, &aux)
}

type AlipayF2FConfig struct {
	AppId       string `json:"app_id"`
	PrivateKey  string `json:"private_key"`
	PublicKey   string `json:"public_key"`
	InvoiceName string `json:"invoice_name"`
	Sandbox     bool   `json:"sandbox"`
}

func (l *AlipayF2FConfig) Marshal() ([]byte, error) {
	type Alias AlipayF2FConfig
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	})
}

func (l *AlipayF2FConfig) Unmarshal(data []byte) error {
	type Alias AlipayF2FConfig
	aux := (*Alias)(l)
	return json.Unmarshal(data, &aux)
}

type EPayConfig struct {
	Pid  string `json:"pid"`
	Url  string `json:"url"`
	Key  string `json:"key"`
	Type string `json:"type"`
}

func (l *EPayConfig) Marshal() ([]byte, error) {
	type Alias EPayConfig
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	})
}

func (l *EPayConfig) Unmarshal(data []byte) error {
	type Alias EPayConfig
	aux := (*Alias)(l)
	return json.Unmarshal(data, &aux)
}

type CryptoSaaSConfig struct {
	Endpoint  string `json:"endpoint"`
	AccountID string `json:"account_id"`
	SecretKey string `json:"secret_key"`
	Type      string `json:"type"`
}

func (l *CryptoSaaSConfig) Marshal() ([]byte, error) {
	type Alias CryptoSaaSConfig
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	})
}

func (l *CryptoSaaSConfig) Unmarshal(data []byte) error {
	type Alias CryptoSaaSConfig
	aux := (*Alias)(l)
	return json.Unmarshal(data, &aux)
}
