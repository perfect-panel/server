package types

const (
	// ForthwithSendEmail forthwith send email
	ForthwithSendEmail = "forthwith:email:send"
)

type (
	SendEmailPayload struct {
		Email   string `json:"to"`
		Subject string `json:"subject"`
		Content string `json:"content"`
	}
)
