package svc

import (
	"context"

	"github.com/perfect-panel/server/internal/model/client"
	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/pkg/device"

	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/model/ads"
	"github.com/perfect-panel/server/internal/model/announcement"
	"github.com/perfect-panel/server/internal/model/auth"
	"github.com/perfect-panel/server/internal/model/coupon"
	"github.com/perfect-panel/server/internal/model/document"
	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/payment"
	"github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/model/system"
	"github.com/perfect-panel/server/internal/model/ticket"
	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/limit"
	"github.com/perfect-panel/server/pkg/nodeMultiplier"
	"github.com/perfect-panel/server/pkg/orm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContext struct {
	DB           *gorm.DB
	Redis        *redis.Client
	Config       config.Config
	Queue        *asynq.Client
	ExchangeRate float64

	//NodeCache   *cache.NodeCacheClient
	AuthModel   auth.Model
	AdsModel    ads.Model
	LogModel    log.Model
	NodeModel   node.Model
	UserModel   user.Model
	OrderModel  order.Model
	ClientModel client.Model
	TicketModel ticket.Model
	//ServerModel        server.Model
	SystemModel       system.Model
	CouponModel       coupon.Model
	PaymentModel      payment.Model
	DocumentModel     document.Model
	SubscribeModel    subscribe.Model
	TrafficLogModel   traffic.Model
	AnnouncementModel announcement.Model

	Restart               func() error
	TelegramBot           *tgbotapi.BotAPI
	NodeMultiplierManager *nodeMultiplier.Manager
	AuthLimiter           *limit.PeriodLimit
	DeviceManager         *device.DeviceManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	// gorm initialize
	db, err := orm.ConnectMysql(orm.Mysql{
		Config: c.MySQL,
	})
	if err != nil {
		panic(err.Error())
	}
	rds := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host,
		Password: c.Redis.Pass,
		DB:       c.Redis.DB,
	})
	err = rds.Ping(context.Background()).Err()
	if err != nil {
		panic(err.Error())
	} else {
		_ = rds.FlushAll(context.Background()).Err()
	}
	authLimiter := limit.NewPeriodLimit(86400, 15, rds, config.SendCountLimitKeyPrefix, limit.Align())
	srv := &ServiceContext{
		DB:           db,
		Redis:        rds,
		Config:       c,
		Queue:        NewAsynqClient(c),
		ExchangeRate: 1.0,
		//NodeCache:   cache.NewNodeCacheClient(rds),
		AuthLimiter: authLimiter,
		AdsModel:    ads.NewModel(db, rds),
		LogModel:    log.NewModel(db),
		NodeModel:   node.NewModel(db, rds),
		AuthModel:   auth.NewModel(db, rds),
		UserModel:   user.NewModel(db, rds),
		OrderModel:  order.NewModel(db, rds),
		ClientModel: client.NewSubscribeApplicationModel(db),
		TicketModel: ticket.NewModel(db, rds),
		//ServerModel:       server.NewModel(db, rds),
		SystemModel:       system.NewModel(db, rds),
		CouponModel:       coupon.NewModel(db, rds),
		PaymentModel:      payment.NewModel(db, rds),
		DocumentModel:     document.NewModel(db, rds),
		SubscribeModel:    subscribe.NewModel(db, rds),
		TrafficLogModel:   traffic.NewModel(db),
		AnnouncementModel: announcement.NewModel(db, rds),
	}
	srv.DeviceManager = NewDeviceManager(srv)
	return srv

}
