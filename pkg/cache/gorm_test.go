package cache

import (
	"context"
	"testing"
	"time"

	"github.com/perfect-panel/server/pkg/orm"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type User struct {
	Id                    int64                 `gorm:"primarykey"`
	Email                 string                `gorm:"index:idx_email;type:varchar(100);unique;not null;comment:电子邮箱"`
	Password              string                `gorm:"type:varchar(100);comment:用户密码;not null"`
	Avatar                string                `gorm:"type:varchar(200);default:'';comment:用户头像"`
	Balance               int64                 `gorm:"default:0;comment:用户余额"`
	Telegram              int64                 `gorm:"default:null;comment:Telegram账号"`
	ReferCode             string                `gorm:"type:varchar(20);default:'';comment:推荐码"`
	RefererId             int64                 `gorm:"comment:推荐人ID"`
	Enable                bool                  `gorm:"default:true;not null;comment:账户是否可用"`
	IsAdmin               bool                  `gorm:"default:false;not null;comment:是否管理员"`
	ValidEmail            bool                  `gorm:"default:false;not null;comment:是否验证邮箱"`
	EnableEmailNotify     bool                  `gorm:"default:false;not null;comment:是否启用邮件通知"`
	EnableTelegramNotify  bool                  `gorm:"default:false;not null;comment:是否启用Telegram通知"`
	EnableBalanceNotify   bool                  `gorm:"default:false;not null;comment:是否启用余额变动通知"`
	EnableLoginNotify     bool                  `gorm:"default:false;not null;comment:是否启用登录通知"`
	EnableSubscribeNotify bool                  `gorm:"default:false;not null;comment:是否启用订阅通知"`
	EnableTradeNotify     bool                  `gorm:"default:false;not null;comment:是否启用交易通知"`
	CreatedAt             time.Time             `gorm:"<-:create;comment:创建时间"`
	UpdatedAt             time.Time             `gorm:"comment:更新时间"`
	DeletedAt             gorm.DeletedAt        `gorm:"default:null;comment:删除时间"`
	IsDel                 soft_delete.DeletedAt `gorm:"softDelete:flag,DeletedAtField:DeletedAt;comment:1:正常 0:删除"` // Use `1` `0` to identify
}

func TestGormCacheCtx(t *testing.T) {
	t.Skipf("skip TestGormCacheCtx test")
	db, err := orm.ConnectMysql(orm.Mysql{
		Config: orm.Config{
			Addr:     "localhost:3306",
			Config:   "charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai",
			Dbname:   "vpnboard",
			Username: "root",
			Password: "mylove520",
		},
	})
	if err != nil {
		t.Error(err)
	}
	rds := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	conn := NewConn(db, rds)
	var u User
	key := "user:id"
	err = conn.QueryCtx(context.Background(), &u, key, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("id = ?", 1).First(v).Error
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("get cache success %+v", u)
}
