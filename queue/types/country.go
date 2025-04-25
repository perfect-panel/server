package types

const (
	// ForthwithGetCountry forthwith country get
	ForthwithGetCountry = "forthwith:country:get"
)

type GetNodeCountry struct {
	Protocol   string `json:"protocol"`
	ServerAddr string `json:"server_addr"`
}
