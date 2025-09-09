package task

import (
	"encoding/json"
	"time"
)

type Type int8

const (
	Undefined Type = -1
	TypeEmail      = iota
	TypeQuota
)

type Task struct {
	Id        int64     `gorm:"primaryKey;autoIncrement;comment:ID"`
	Type      int8      `gorm:"not null;comment:Task Type"`
	Scope     string    `gorm:"type:text;comment:Task Scope"`
	Content   string    `gorm:"type:text;comment:Task Content"`
	Status    int8      `gorm:"not null;default:0;comment:Task Status: 0: Pending, 1: In Progress, 2: Completed, 3: Failed"`
	Errors    string    `gorm:"type:text;comment:Task Errors"`
	Total     uint64    `gorm:"column:total;not null;default:0;comment:Total Number"`
	Current   uint64    `gorm:"column:current;not null;default:0;comment:Current Number"`
	CreatedAt time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt time.Time `gorm:"comment:Update Time"`
}

func (Task) TableName() string {
	return "task"
}

type ScopeType int8

const (
	ScopeAll     ScopeType = iota + 1 // All users
	ScopeActive                       // Active users
	ScopeExpired                      // Expired users
	ScopeNone                         // No Subscribe
	ScopeSkip                         // Skip user filtering
)

func (t ScopeType) Int8() int8 {
	return int8(t)
}

type EmailScope struct {
	Type              int8     `gorm:"not null;comment:Scope Type"`
	RegisterStartTime int64    `json:"register_start_time"`
	RegisterEndTime   int64    `json:"register_end_time"`
	Recipients        []string `json:"recipients"` // list of email addresses
	Additional        []string `json:"additional"` // additional email addresses
	Scheduled         int64    `json:"scheduled"`  // scheduled time (unix timestamp)
	Interval          uint8    `json:"interval"`   // interval in seconds
	Limit             uint64   `json:"limit"`      // daily send limit
}

func (s *EmailScope) Marshal() ([]byte, error) {
	type Alias EmailScope
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}

func (s *EmailScope) Unmarshal(data []byte) error {
	type Alias EmailScope
	aux := (*Alias)(s)
	return json.Unmarshal(data, &aux)
}

type EmailContent struct {
	Subject string `json:"subject"`
	Content string `json:"content"`
}

func (c *EmailContent) Marshal() ([]byte, error) {
	type Alias EmailContent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	})
}

func (c *EmailContent) Unmarshal(data []byte) error {
	type Alias EmailContent
	aux := (*Alias)(c)
	return json.Unmarshal(data, &aux)
}

type QuotaScope struct {
	Type              int8    `gorm:"not null;comment:Scope Type"`
	RegisterStartTime int64   `json:"register_start_time"`
	RegisterEndTime   int64   `json:"register_end_time"`
	Recipients        []int64 `json:"recipients"` // list of user subs IDs
}

func (s *QuotaScope) Marshal() ([]byte, error) {
	type Alias QuotaScope
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}

func (s *QuotaScope) Unmarshal(data []byte) error {
	type Alias QuotaScope
	aux := (*Alias)(s)
	return json.Unmarshal(data, &aux)
}

type QuotaType int8

const (
	QuotaTypeReset QuotaType = iota + 1 // Reset Subscribe  Quota
	QuotaTypeDays                       // Add Subscribe Days
	QuotaTypeGift                       // Add Gift Amount
)

type QuotaContent struct {
	Type int8   `json:"type"`
	Days uint64 `json:"days,omitempty"` // days to add
	Gift uint8  `json:"gift,omitempty"` // Invoice amount ratio(%) to gift amount
}

func ParseScopeType(t int8) ScopeType {
	switch t {
	case 1:
		return ScopeAll
	case 2:
		return ScopeActive
	case 3:
		return ScopeExpired
	case 4:
		return ScopeNone
	default:
		return ScopeSkip
	}
}
