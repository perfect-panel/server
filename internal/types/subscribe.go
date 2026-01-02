package types

type (
	SubscribeRequest struct {
		Flag   string
		Token  string
		Type   string
		UA     string
		Params map[string]string
	}
	SubscribeResponse struct {
		Config []byte
		Header string
	}
)
