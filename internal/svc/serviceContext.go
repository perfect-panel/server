package svc

import (
	"context"

	"github.com/perfect-panel/server/internal/model/client"
	"github.com/perfect-panel/server/pkg/device"

	"github.com/perfect-panel/server/internal/model/ads"
	"github.com/perfect-panel/server/internal/model/cache"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/model/announcement"
	"github.com/perfect-panel/server/internal/model/application"
	"github.com/perfect-panel/server/internal/model/auth"
	"github.com/perfect-panel/server/internal/model/coupon"
	"github.com/perfect-panel/server/internal/model/document"
	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/payment"
	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/model/subscribeType"
	"github.com/perfect-panel/server/internal/model/system"
	"github.com/perfect-panel/server/internal/model/ticket"
	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/limit"
	"github.com/perfect-panel/server/pkg/nodeMultiplier"
	"github.com/perfect-panel/server/pkg/orm"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContext struct {
	DB                 *gorm.DB
	Redis              *redis.Client
	Config             config.Config
	Queue              *asynq.Client
	NodeCache          *cache.NodeCacheClient
	AuthModel          auth.Model
	AdsModel           ads.Model
	LogModel           log.Model
	UserModel          user.Model
	OrderModel         order.Model
	ClientModel        client.Model
	TicketModel        ticket.Model
	ServerModel        server.Model
	SystemModel        system.Model
	CouponModel        coupon.Model
	PaymentModel       payment.Model
	DocumentModel      document.Model
	SubscribeModel     subscribe.Model
	TrafficLogModel    traffic.Model
	ApplicationModel   application.Model
	AnnouncementModel  announcement.Model
	SubscribeTypeModel subscribeType.Model

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
		DB:                db,
		Redis:             rds,
		Config:            c,
		Queue:             NewAsynqClient(c),
		NodeCache:         cache.NewNodeCacheClient(rds),
		AuthLimiter:       authLimiter,
		AdsModel:          ads.NewModel(db, rds),
		LogModel:          log.NewModel(db),
		AuthModel:         auth.NewModel(db, rds),
		UserModel:         user.NewModel(db, rds),
		OrderModel:        order.NewModel(db, rds),
		ClientModel:       client.NewSubscribeApplicationModel(db),
		TicketModel:       ticket.NewModel(db, rds),
		ServerModel:       server.NewModel(db, rds),
		SystemModel:       system.NewModel(db, rds),
		CouponModel:       coupon.NewModel(db, rds),
		PaymentModel:      payment.NewModel(db, rds),
		DocumentModel:     document.NewModel(db, rds),
		SubscribeModel:    subscribe.NewModel(db, rds),
		TrafficLogModel:   traffic.NewModel(db),
		ApplicationModel:  application.NewModel(db, rds),
		AnnouncementModel: announcement.NewModel(db, rds),
	}
	srv.DeviceManager = NewDeviceManager(srv)
	return srv

}
