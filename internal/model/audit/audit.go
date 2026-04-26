package audit

// V4.3 audit_log — 重置/加购/删除/超额 全埋点(决策 13/35)。
// 单文件 model:简单 Insert + 分页查询足够覆盖管理端需求,管理端列表在 Phase 8 实现。

import (
	"context"
	"encoding/json"
	"time"

	"github.com/perfect-panel/server/pkg/logger"
	"gorm.io/gorm"
)

// Actor 标识动作发起方。
const (
	ActorUser   = "user"
	ActorAdmin  = "admin"
	ActorSystem = "system"
)

// Action 关键动作 key。新增请保持小写下划线,前端分类靠匹配前缀。
const (
	ActionPurchase        = "purchase"
	ActionRenew           = "renew"
	ActionAddDevice       = "add_device"
	ActionAddonTraffic    = "addon_traffic"
	ActionResetDevice     = "reset_device"
	ActionResetAllDevices = "reset_all_devices"
	ActionDisableDevice   = "disable_device"
	ActionEnableDevice    = "enable_device"
	ActionRenameDevice    = "rename_device"
	ActionThrottleStart   = "throttle_start"
	ActionThrottleCutOff  = "throttle_cut_off"
	ActionAdminLoginRemote = "admin_login_remote"
)

type AuditLog struct {
	Id        int64     `gorm:"primaryKey"`
	UserId    int64     `gorm:"column:user_id;index:idx_user_action,priority:1;not null;default:0"`
	Actor     string    `gorm:"column:actor;type:varchar(16);index:idx_actor,priority:1;not null;default:''"`
	ActorId   int64     `gorm:"column:actor_id;index:idx_actor,priority:2;not null;default:0"`
	Action    string    `gorm:"column:action;type:varchar(64);index:idx_user_action,priority:2;not null;default:''"`
	Target    string    `gorm:"column:target;type:varchar(255);not null;default:''"`
	Detail    string    `gorm:"column:detail;type:text"`
	ClientIp  string    `gorm:"column:client_ip;type:varchar(45);not null;default:''"`
	CreatedAt time.Time `gorm:"<-:create;index:idx_created_at"`
}

func (*AuditLog) TableName() string {
	return "audit_log"
}

type Model interface {
	Append(ctx context.Context, entry *AuditLog, tx ...*gorm.DB) error
	AppendDetail(ctx context.Context, entry *AuditLog, detail interface{}, tx ...*gorm.DB) error
	Query(ctx context.Context, filter *Filter, page, size int) ([]*AuditLog, int64, error)
	CleanOlderThan(ctx context.Context, before time.Time) (int64, error)
}

type Filter struct {
	UserId  *int64
	Actor   string
	ActorId *int64
	Action  string
	Since   *time.Time
	Until   *time.Time
}

type defaultModel struct {
	db *gorm.DB
}

func NewModel(db *gorm.DB) Model {
	return &defaultModel{db: db}
}

func (m *defaultModel) Append(ctx context.Context, entry *AuditLog, tx ...*gorm.DB) error {
	conn := m.db.WithContext(ctx)
	if len(tx) > 0 {
		conn = tx[0]
	}
	if err := conn.Create(entry).Error; err != nil {
		// audit 失败不应阻断主业务,降级到日志。
		logger.Errorf("[audit] append failed: action=%s user_id=%d err=%v", entry.Action, entry.UserId, err)
		return err
	}
	return nil
}

func (m *defaultModel) AppendDetail(ctx context.Context, entry *AuditLog, detail interface{}, tx ...*gorm.DB) error {
	if detail != nil {
		b, err := json.Marshal(detail)
		if err == nil {
			entry.Detail = string(b)
		}
	}
	return m.Append(ctx, entry, tx...)
}

func (m *defaultModel) Query(ctx context.Context, filter *Filter, page, size int) ([]*AuditLog, int64, error) {
	conn := m.db.WithContext(ctx).Model(&AuditLog{})
	if filter != nil {
		if filter.UserId != nil {
			conn = conn.Where("user_id = ?", *filter.UserId)
		}
		if filter.Actor != "" {
			conn = conn.Where("actor = ?", filter.Actor)
		}
		if filter.ActorId != nil {
			conn = conn.Where("actor_id = ?", *filter.ActorId)
		}
		if filter.Action != "" {
			conn = conn.Where("action = ?", filter.Action)
		}
		if filter.Since != nil {
			conn = conn.Where("created_at >= ?", *filter.Since)
		}
		if filter.Until != nil {
			conn = conn.Where("created_at < ?", *filter.Until)
		}
	}
	var total int64
	if err := conn.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*AuditLog
	if size <= 0 {
		size = 20
	}
	if page <= 0 {
		page = 1
	}
	err := conn.Order("created_at DESC").
		Limit(size).Offset((page - 1) * size).
		Find(&list).Error
	return list, total, err
}

func (m *defaultModel) CleanOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res := m.db.WithContext(ctx).Where("created_at < ?", before).Delete(&AuditLog{})
	return res.RowsAffected, res.Error
}
