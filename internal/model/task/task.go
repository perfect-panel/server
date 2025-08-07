package task

import "time"

type EmailTask struct {
	Id                int64     `gorm:"column:id;primaryKey;autoIncrement;comment:ID"`
	Subject           string    `gorm:"column:subject;type:varchar(255);not null;comment:Email Subject"`
	Content           string    `gorm:"column:content;type:text;not null;comment:Email Content"`
	Recipients        string    `gorm:"column:recipient;type:text;not null;comment:Email Recipient"`
	Scope             string    `gorm:"column:scope;type:varchar(50);not null;comment:Email Scope"`
	RegisterStartTime time.Time `gorm:"column:register_start_time;default:null;comment:Register Start Time"`
	RegisterEndTime   time.Time `gorm:"column:register_end_time;default:null;comment:Register End Time"`
	Additional        string    `gorm:"column:additional;type:text;default:null;comment:Additional Information"`
	Scheduled         time.Time `gorm:"column:scheduled;not null;comment:Scheduled Time"`
	Interval          uint8     `gorm:"column:interval;not null;comment:Interval in Seconds"`
	Limit             uint64    `gorm:"column:limit;not null;comment:Daily send limit"`
	Status            uint8     `gorm:"column:status;not null;comment:Daily Status"`
	Errors            string    `gorm:"column:errors;type:text;not null;comment:Errors"`
	Total             uint64    `gorm:"column:total;not null;default:0;comment:Total Number"`
	Current           uint64    `gorm:"column:current;not null;default:0;comment:Current Number"`
	CreatedAt         time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt         time.Time `gorm:"comment:Update Time"`
}

func (EmailTask) TableName() string {
	return "email_task"
}
