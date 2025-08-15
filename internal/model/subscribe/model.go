package subscribe

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

//	type Details struct {
//		Id                   int64  `gorm:"primaryKey"`
//		Name                 string `gorm:"type:varchar(255);not null;default:'';comment:Subscribe Name"`
//		Description          string `gorm:"type:text;comment:Subscribe Description"`
//		UnitPrice            int64  `gorm:"type:int;not null;default:0;comment:Unit Price"`
//		UnitTime             string `gorm:"type:varchar(255);not null;default:'';comment:Unit Time"`
//		Discount             string `gorm:"type:text;comment:Discount"`
//		Replacement          int64  `gorm:"type:int;not null;default:0;comment:Replacement"`
//		Inventory            int64  `gorm:"type:int;not null;default:0;comment:Inventory"`
//		Traffic              int64  `gorm:"type:int;not null;default:0;comment:Traffic"`
//		SpeedLimit           int64  `gorm:"type:int;not null;default:0;comment:Speed Limit"`
//		DeviceLimit          int64  `gorm:"type:int;not null;default:0;comment:Device Limit"`
//		GroupId              int64  `gorm:"type:bigint;comment:Group Id"`
//		Quota                int64  `gorm:"type:int;not null;default:0;comment:Quota"`
//		Show                 *bool  `gorm:"type:tinyint(1);not null;default:0;comment:Show"`
//		Sell                 *bool  `gorm:"type:tinyint(1);not null;default:0;comment:Sell"`
//		DeductionRatio       int64  `gorm:"type:int;default:0;comment:Deduction Ratio"`
//		PurchaseWithDiscount bool   `gorm:"type:tinyint(1);default:0;comment:PurchaseWithDiscount"`
//		ResetCycle           int64  `gorm:"type:int;default:0;comment:Reset Cycle"`
//		RenewalReset         bool   `gorm:"type:tinyint(1);default:0;comment:Renew Reset"`
//	}
type customSubscribeLogicModel interface {
	QuerySubscribeListByPage(ctx context.Context, page, size int, group int64, search string) (total int64, list []*Subscribe, err error)
	QuerySubscribeList(ctx context.Context) ([]*Subscribe, error)
	QuerySubscribeListByShow(ctx context.Context) ([]*Subscribe, error)
	QuerySubscribeIdsByServerIdAndServerGroupId(ctx context.Context, serverId, serverGroupId int64) ([]*Subscribe, error)
	QuerySubscribeMinSortByIds(ctx context.Context, ids []int64) (int64, error)
	QuerySubscribeListByIds(ctx context.Context, ids []int64) ([]*Subscribe, error)
	ClearCache(ctx context.Context, id ...int64) error
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customSubscribeModel{
		defaultSubscribeModel: newSubscribeModel(conn, c),
	}
}

// QuerySubscribeListByPage  Get Subscribe List
func (m *customSubscribeModel) QuerySubscribeListByPage(ctx context.Context, page, size int, group int64, search string) (total int64, list []*Subscribe, err error) {
	err = m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		// About to be abandoned
		_ = conn.Model(&Subscribe{}).
			Where("sort = ?", 0).
			Update("sort", gorm.Expr("id"))

		conn = conn.Model(&Subscribe{})
		if group > 0 {
			conn = conn.Where("group_id = ?", group)
		}
		if search != "" {
			conn = conn.Where("`name` like ? or `description` like ?", "%"+search+"%", "%"+search+"%")
		}
		return conn.Count(&total).Order("sort ASC").Limit(size).Offset((page - 1) * size).Find(v).Error
	})
	return total, list, err
}

// QuerySubscribeList Get Subscribe List
func (m *customSubscribeModel) QuerySubscribeList(ctx context.Context) ([]*Subscribe, error) {
	var list []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.Model(&Subscribe{})
		return conn.Where("`sell` = true").Order("sort ").Find(v).Error
	})
	return list, err
}

func (m *customSubscribeModel) QuerySubscribeIdsByServerIdAndServerGroupId(ctx context.Context, serverId, serverGroupId int64) ([]*Subscribe, error) {
	var data []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("FIND_IN_SET(?, server)", serverId).Or("FIND_IN_SET(?, server_group)", serverGroupId).Find(v).Error
	})
	return data, err
}

// QuerySubscribeListByShow Get Subscribe List By Show
func (m *customSubscribeModel) QuerySubscribeListByShow(ctx context.Context) ([]*Subscribe, error) {
	var list []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.Model(&Subscribe{})
		return conn.Where("`show` = true").Find(v).Error
	})
	return list, err
}

func (m *customSubscribeModel) QuerySubscribeMinSortByIds(ctx context.Context, ids []int64) (int64, error) {
	var minSort int64
	err := m.QueryNoCacheCtx(ctx, &minSort, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("id IN ?", ids).Select("COALESCE(MIN(sort), 0)").Scan(v).Error
	})
	return minSort, err
}

func (m *customSubscribeModel) QuerySubscribeListByIds(ctx context.Context, ids []int64) ([]*Subscribe, error) {
	var list []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("id IN ?", ids).Find(v).Error
	})
	return list, err
}

func (m *customSubscribeModel) ClearCache(ctx context.Context, ids ...int64) error {
	if len(ids) <= 0 {
		return nil
	}

	var cacheKeys []string
	for _, id := range ids {
		data, err := m.FindOne(ctx, id)
		if err != nil {
			return err
		}
		cacheKeys = append(cacheKeys, m.getCacheKeys(data)...)
	}
	return m.CachedConn.DelCacheCtx(ctx, cacheKeys...)
}
