package surfboard

import "time"

type UserInfo struct {
	UUID         string
	Upload       int64
	Download     int64
	TotalTraffic int64
	ExpiredDate  time.Time
	SubscribeURL string
}
