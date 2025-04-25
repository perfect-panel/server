package cache

type NodeOnlineUser struct {
	SID int64
	IP  string
}

type NodeTodayTrafficRank struct {
	ID       int64
	Name     string
	Upload   int64
	Download int64
	Total    int64
}

type UserTodayTrafficRank struct {
	SID      int64
	Upload   int64
	Download int64
	Total    int64
}

type UserTraffic struct {
	UID      int64
	Upload   int64
	Download int64
}

type NodeStatus struct {
	Cpu       float64
	Mem       float64
	Disk      float64
	UpdatedAt int64
}
