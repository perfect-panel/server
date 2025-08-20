package types

const (
	// ForthwithSendSms forthwith send email
	ForthwithSendSms = "forthwith:sms:send"
)

type (
	SendSmsPayload struct {
		Type          uint8  `json:"type"`
		Telephone     string `json:"telephone"`
		TelephoneArea string `json:"area"`
		Content       string `json:"content"`
	}
)
