package repository

import (
	"context"

	"github.com/perfect-panel/server/internal/model/ads"
	"github.com/perfect-panel/server/internal/model/announcement"
	"github.com/perfect-panel/server/internal/model/auth"
	"github.com/perfect-panel/server/internal/model/client"
	"github.com/perfect-panel/server/internal/model/coupon"
	"github.com/perfect-panel/server/internal/model/document"
	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/payment"
	"github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/model/system"
	"github.com/perfect-panel/server/internal/model/ticket"
	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Store interface {
	Ads() ads.Model
	Announcement() announcement.Model
	Auth() auth.Model
	Client() client.Model
	Coupon() coupon.Model
	Document() document.Model
	Log() log.Model
	Node() node.Model
	Order() order.Model
	Payment() payment.Model
	Subscribe() subscribe.Model
	System() system.Model
	Ticket() ticket.Model
	TrafficLog() traffic.Model
	User() user.Model

	InTx(ctx context.Context, fn func(store Store) error) error
}

var _ Store = (*GormStore)(nil)

type GormStore struct {
	db    *gorm.DB
	redis *redis.Client

	ads          ads.Model
	announcement announcement.Model
	auth         auth.Model
	client       client.Model
	coupon       coupon.Model
	document     document.Model
	log          log.Model
	node         node.Model
	order        order.Model
	payment      payment.Model
	subscribe    subscribe.Model
	system       system.Model
	ticket       ticket.Model
	trafficLog   traffic.Model
	user         user.Model
}

func NewGormStore(db *gorm.DB, rds *redis.Client) *GormStore {
	return &GormStore{
		db:           db,
		redis:        rds,
		ads:          ads.NewModel(db, rds),
		announcement: announcement.NewModel(db, rds),
		auth:         auth.NewModel(db, rds),
		client:       client.NewSubscribeApplicationModel(db),
		coupon:       coupon.NewModel(db, rds),
		document:     document.NewModel(db, rds),
		log:          log.NewModel(db),
		node:         node.NewModel(db, rds),
		order:        order.NewModel(db, rds),
		payment:      payment.NewModel(db, rds),
		subscribe:    subscribe.NewModel(db, rds),
		system:       system.NewModel(db, rds),
		ticket:       ticket.NewModel(db, rds),
		trafficLog:   traffic.NewModel(db),
		user:         user.NewModel(db, rds),
	}
}

func (s *GormStore) Ads() ads.Model {
	return s.ads
}

func (s *GormStore) Announcement() announcement.Model {
	return s.announcement
}

func (s *GormStore) Auth() auth.Model {
	return s.auth
}

func (s *GormStore) Client() client.Model {
	return s.client
}

func (s *GormStore) Coupon() coupon.Model {
	return s.coupon
}

func (s *GormStore) Document() document.Model {
	return s.document
}

func (s *GormStore) Log() log.Model {
	return s.log
}

func (s *GormStore) Node() node.Model {
	return s.node
}

func (s *GormStore) Order() order.Model {
	return s.order
}

func (s *GormStore) Payment() payment.Model {
	return s.payment
}

func (s *GormStore) Subscribe() subscribe.Model {
	return s.subscribe
}

func (s *GormStore) System() system.Model {
	return s.system
}

func (s *GormStore) Ticket() ticket.Model {
	return s.ticket
}

func (s *GormStore) TrafficLog() traffic.Model {
	return s.trafficLog
}

func (s *GormStore) User() user.Model {
	return s.user
}

func (s *GormStore) InTx(ctx context.Context, fn func(store Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewGormStore(tx, s.redis))
	})
}
