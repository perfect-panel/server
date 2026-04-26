package types

const (
	// ForthwithSendEmail forthwith send email
	ForthwithSendEmail = "forthwith:email:send"
)

const (
	EmailTypeVerify        = "verify"
	EmailTypeMaintenance   = "maintenance"
	EmailTypeExpiration    = "expiration"
	EmailTypeTrafficExceed = "traffic_exceed"
	EmailTypeCustom        = "custom"
	// V4.3 通知模板(决策 20 + 7.1 通知矩阵)
	// Content: { Subject string, Body string, ... user vars } — Body 已渲染
	EmailTypeNotice = "notice"
)


type (
	SendEmailPayload struct {
		Type    string                 `json:"type"`
		Email   string                 `json:"to"`
		Subject string                 `json:"subject"`
		Content map[string]interface{} `json:"content"`
	}
)
