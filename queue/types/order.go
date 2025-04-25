package types

const (
	DeferCloseOrder        = "defer:order:close"
	ForthwithActivateOrder = "forthwith:order:activate"
)

type (
	DeferCloseOrderPayload struct {
		OrderNo string `json:"order_no"`
	}
	ForthwithActivateOrderPayload struct {
		OrderNo string `json:"order_no"`
	}
)
