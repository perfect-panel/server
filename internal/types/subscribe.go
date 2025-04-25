package types

type (
	SubscribeRequest struct {
		Flag  string
		Token string
		UA    string
	}
	SubscribeResponse struct {
		Config []byte
		Header string
	}
)
