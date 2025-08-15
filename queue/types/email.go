package types

const (
	// ForthwithSendEmail forthwith send email
	ForthwithSendEmail = "forthwith:email:send"
	// ScheduledBatchSendEmail scheduled batch send email
	ScheduledBatchSendEmail = "scheduled:email:batch"
)

type (
	SendEmailPayload struct {
		Type    string `json:"type"`
		Email   string `json:"to"`
		Subject string `json:"subject"`
		Content string `json:"content"`
	}
)
