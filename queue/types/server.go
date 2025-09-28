package types

const ForthwithTrafficStatistics = "forthwith:traffic:statistics"

type UserTraffic struct {
	SID      int64 `json:"uid"`
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
}

type TrafficStatistics struct {
	ServerId int64         `json:"server_id"`
	Protocol string        `json:"protocol"`
	Logs     []UserTraffic `json:"logs"`
}

type NodeStatus struct {
	OnlineUsers []OnlineUser `json:"online_users"`
	Status      ServerStatus `json:"status"`
	LastAt      int64        `json:"last_at"`
}

type OnlineUser struct {
	UID int64  `json:"uid"`
	IP  string `json:"ip"`
}

type ServerStatus struct {
	Cpu       float64 `json:"cpu"`
	Mem       float64 `json:"mem"`
	Disk      float64 `json:"disk"`
	UpdatedAt int64   `json:"updated_at"`
}

type ServerTrafficCount struct {
	ServerId  int64  `json:"server_id"`
	Name      string `json:"name"`
	Today     int64  `json:"today"`
	Yesterday int64  `json:"yesterday"`
}
