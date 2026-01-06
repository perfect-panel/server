package redemption

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ RedemptionCodeModel = (*customRedemptionCodeModel)(nil)
var _ RedemptionRecordModel = (*customRedemptionRecordModel)(nil)

var (
	cacheRedemptionCodeIdPrefix   = "cache:redemption_code:id:"
	cacheRedemptionCodeCodePrefix = "cache:redemption_code:code:"
	cacheRedemptionRecordIdPrefix = "cache:redemption_record:id:"
)

type (
	RedemptionCodeModel interface {
		Insert(ctx context.Context, data *RedemptionCode) error
		FindOne(ctx context.Context, id int64) (*RedemptionCode, error)
		FindOneByCode(ctx context.Context, code string) (*RedemptionCode, error)
		Update(ctx context.Context, data *RedemptionCode) error
		Delete(ctx context.Context, id int64) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
		customRedemptionCodeLogicModel
	}

	RedemptionRecordModel interface {
		Insert(ctx context.Context, data *RedemptionRecord) error
		FindOne(ctx context.Context, id int64) (*RedemptionRecord, error)
		Update(ctx context.Context, data *RedemptionRecord) error
		Delete(ctx context.Context, id int64) error
		customRedemptionRecordLogicModel
	}

	customRedemptionCodeLogicModel interface {
		QueryRedemptionCodeListByPage(ctx context.Context, page, size int, subscribePlan int64, unitTime string, code string) (total int64, list []*RedemptionCode, err error)
		BatchDelete(ctx context.Context, ids []int64) error
		IncrementUsedCount(ctx context.Context, id int64) error
	}

	customRedemptionRecordLogicModel interface {
		QueryRedemptionRecordListByPage(ctx context.Context, page, size int, userId int64, codeId int64) (total int64, list []*RedemptionRecord, err error)
		FindByUserId(ctx context.Context, userId int64) ([]*RedemptionRecord, error)
		FindByCodeId(ctx context.Context, codeId int64) ([]*RedemptionRecord, error)
	}

	customRedemptionCodeModel struct {
		*defaultRedemptionCodeModel
	}
	defaultRedemptionCodeModel struct {
		cache.CachedConn
		table string
	}

	customRedemptionRecordModel struct {
		*defaultRedemptionRecordModel
	}
	defaultRedemptionRecordModel struct {
		cache.CachedConn
		table string
	}
)

func newRedemptionCodeModel(db *gorm.DB, c *redis.Client) *defaultRedemptionCodeModel {
	return &defaultRedemptionCodeModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`redemption_code`",
	}
}

func newRedemptionRecordModel(db *gorm.DB, c *redis.Client) *defaultRedemptionRecordModel {
	return &defaultRedemptionRecordModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`redemption_record`",
	}
}

// RedemptionCode cache methods
func (m *defaultRedemptionCodeModel) getCacheKeys(data *RedemptionCode) []string {
	if data == nil {
		return []string{}
	}
	codeIdKey := fmt.Sprintf("%s%v", cacheRedemptionCodeIdPrefix, data.Id)
	codeCodeKey := fmt.Sprintf("%s%v", cacheRedemptionCodeCodePrefix, data.Code)
	cacheKeys := []string{
		codeIdKey,
		codeCodeKey,
	}
	return cacheKeys
}

func (m *defaultRedemptionCodeModel) Insert(ctx context.Context, data *RedemptionCode) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultRedemptionCodeModel) FindOne(ctx context.Context, id int64) (*RedemptionCode, error) {
	codeIdKey := fmt.Sprintf("%s%v", cacheRedemptionCodeIdPrefix, id)
	var resp RedemptionCode
	err := m.QueryCtx(ctx, &resp, codeIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&RedemptionCode{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultRedemptionCodeModel) FindOneByCode(ctx context.Context, code string) (*RedemptionCode, error) {
	codeCodeKey := fmt.Sprintf("%s%v", cacheRedemptionCodeCodePrefix, code)
	var resp RedemptionCode
	err := m.QueryCtx(ctx, &resp, codeCodeKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&RedemptionCode{}).Where("`code` = ?", code).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultRedemptionCodeModel) Update(ctx context.Context, data *RedemptionCode) error {
	old, err := m.FindOne(ctx, data.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Save(data).Error
	}, m.getCacheKeys(old)...)
	return err
}

func (m *defaultRedemptionCodeModel) Delete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Delete(&RedemptionCode{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultRedemptionCodeModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}

// RedemptionCode custom logic methods
func (m *customRedemptionCodeModel) QueryRedemptionCodeListByPage(ctx context.Context, page, size int, subscribePlan int64, unitTime string, code string) (total int64, list []*RedemptionCode, err error) {
	err = m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		db := conn.Model(&RedemptionCode{})
		if subscribePlan != 0 {
			db = db.Where("subscribe_plan = ?", subscribePlan)
		}
		if unitTime != "" {
			db = db.Where("unit_time = ?", unitTime)
		}
		if code != "" {
			db = db.Where("code like ?", "%"+code+"%")
		}
		return db.Count(&total).Limit(size).Offset((page - 1) * size).Order("created_at DESC").Find(v).Error
	})
	return total, list, err
}

func (m *customRedemptionCodeModel) BatchDelete(ctx context.Context, ids []int64) error {
	var err error
	for _, id := range ids {
		if err = m.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (m *customRedemptionCodeModel) IncrementUsedCount(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	data.UsedCount++
	return m.Update(ctx, data)
}

// RedemptionRecord cache methods
func (m *defaultRedemptionRecordModel) getCacheKeys(data *RedemptionRecord) []string {
	if data == nil {
		return []string{}
	}
	recordIdKey := fmt.Sprintf("%s%v", cacheRedemptionRecordIdPrefix, data.Id)
	cacheKeys := []string{
		recordIdKey,
	}
	return cacheKeys
}

func (m *defaultRedemptionRecordModel) Insert(ctx context.Context, data *RedemptionRecord) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultRedemptionRecordModel) FindOne(ctx context.Context, id int64) (*RedemptionRecord, error) {
	recordIdKey := fmt.Sprintf("%s%v", cacheRedemptionRecordIdPrefix, id)
	var resp RedemptionRecord
	err := m.QueryCtx(ctx, &resp, recordIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&RedemptionRecord{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultRedemptionRecordModel) Update(ctx context.Context, data *RedemptionRecord) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Save(data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultRedemptionRecordModel) Delete(ctx context.Context, id int64) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Delete(&RedemptionRecord{}, id).Error
	}, m.getCacheKeys(nil)...)
	return err
}

// RedemptionRecord custom logic methods
func (m *customRedemptionRecordModel) QueryRedemptionRecordListByPage(ctx context.Context, page, size int, userId int64, codeId int64) (total int64, list []*RedemptionRecord, err error) {
	err = m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		db := conn.Model(&RedemptionRecord{})
		if userId != 0 {
			db = db.Where("user_id = ?", userId)
		}
		if codeId != 0 {
			db = db.Where("redemption_code_id = ?", codeId)
		}
		return db.Count(&total).Limit(size).Offset((page - 1) * size).Order("created_at DESC").Find(v).Error
	})
	return total, list, err
}

func (m *customRedemptionRecordModel) FindByUserId(ctx context.Context, userId int64) ([]*RedemptionRecord, error) {
	var list []*RedemptionRecord
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&RedemptionRecord{}).Where("user_id = ?", userId).Order("created_at DESC").Find(v).Error
	})
	return list, err
}

func (m *customRedemptionRecordModel) FindByCodeId(ctx context.Context, codeId int64) ([]*RedemptionRecord, error) {
	var list []*RedemptionRecord
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&RedemptionRecord{}).Where("redemption_code_id = ?", codeId).Order("created_at DESC").Find(v).Error
	})
	return list, err
}
