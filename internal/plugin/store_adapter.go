package plugin

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// modelWhitelist 定义插件可查询的数据库表白名单
var modelWhitelist = map[string]string{
	"ads":                       "ads",
	"announcement":              "announcement",
	"auth_method":               "auth_method",
	"coupon":                    "coupon",
	"document":                  "document",
	"node":                      "nodes",
	"order":                     "order",
	"payment":                   "payment",
	"server":                    "servers",
	"subscribe":                 "subscribe",
	"subscribe_application":     "subscribe_application",
	"subscribe_group":           "subscribe_group",
	"system":                    "system",
	"system_log":                "system_logs",
	"task":                      "task",
	"ticket":                    "ticket",
	"ticket_follow":             "ticket_follow",
	"traffic_log":               "traffic_log",
	"user":                      "user",
	"user_auth_method":          "user_auth_methods",
	"user_device":               "user_device",
	"user_device_online_record": "user_device_online_record",
	"user_subscribe":            "user_subscribe",
	"user_withdrawal":           "user_withdrawal",
}

var modelFieldWhitelist = map[string]map[string]bool{
	"ads": allowFields(
		"id", "title", "type", "content", "description", "target_url", "start_time",
		"end_time", "status", "created_at", "updated_at",
	),
	"announcement": allowFields(
		"id", "title", "content", "show", "pinned", "popup", "created_at", "updated_at",
	),
	"auth_method": allowFields(
		"id", "method", "enabled", "created_at", "updated_at",
	),
	"coupon": allowFields(
		"id", "name", "code", "count", "type", "discount", "start_time", "expire_time",
		"user_limit", "subscribe", "used_count", "enable", "created_at", "updated_at",
	),
	"document": allowFields(
		"id", "title", "content", "tags", "show", "created_at", "updated_at",
	),
	"node": allowFields(
		"id", "name", "tags", "port", "address", "server_id", "protocol", "enabled",
		"sort", "created_at", "updated_at",
	),
	"order": allowFields(
		"id", "parent_id", "user_id", "order_no", "type", "quantity", "price", "amount",
		"gift_amount", "discount", "coupon", "coupon_discount", "commission", "payment_id",
		"method", "fee_amount", "trade_no", "status", "subscribe_id", "is_new",
		"created_at", "updated_at",
	),
	"payment": allowFields(
		"id", "name", "platform", "icon", "domain", "description", "fee_mode",
		"fee_percent", "fee_amount", "sort", "enable",
	),
	"server": allowFields(
		"id", "name", "country", "city", "address", "sort", "last_reported_at",
		"created_at", "updated_at",
	),
	"subscribe": allowFields(
		"id", "name", "language", "description", "unit_price", "unit_time", "discount",
		"replacement", "inventory", "traffic", "speed_limit", "device_limit", "quota",
		"nodes", "node_tags", "show", "sell", "sort", "deduction_ratio",
		"allow_deduction", "reset_cycle", "renewal_reset", "show_original_price",
		"created_at", "updated_at",
	),
	"subscribe_application": allowFields(
		"id", "name", "icon", "description", "scheme", "user_agent", "is_default",
		"subscribe_template", "output_format", "download_link", "created_at", "updated_at",
	),
	"subscribe_group": allowFields(
		"id", "name", "description", "created_at", "updated_at",
	),
	"system": allowFields(
		"id", "category", "key", "type", "desc", "created_at", "updated_at",
	),
	"system_log": allowFields(
		"id", "type", "date", "object_id", "created_at",
	),
	"task": allowFields(
		"id", "type", "scope", "content", "status", "errors", "total", "current",
		"created_at", "updated_at",
	),
	"ticket": allowFields(
		"id", "user_id", "status", "title", "description", "created_at", "updated_at",
	),
	"ticket_follow": allowFields(
		"id", "ticket_id", "from", "type", "content", "created_at",
	),
	"ticket_reply": allowFields(
		"ticket_id", "from", "type", "content",
	),
	"traffic_log": allowFields(
		"id", "server_id", "user_id", "subscribe_id", "download", "upload", "timestamp",
	),
	"user": allowFields(
		"id", "avatar", "balance", "refer_code", "referer_id", "commission",
		"referral_percentage", "only_first_purchase", "gift_amount", "enable", "is_admin",
		"enable_balance_notify", "enable_login_notify", "enable_subscribe_notify",
		"enable_trade_notify", "rules", "created_at", "updated_at",
	),
	"user_auth_method": allowFields(
		"id", "user_id", "auth_type", "auth_identifier", "verified", "created_at", "updated_at",
	),
	"user_device": allowFields(
		"id", "ip", "user_id", "user_agent", "identifier", "online", "enabled",
		"created_at", "updated_at",
	),
	"user_device_online_record": allowFields(
		"id", "user_id", "identifier", "online_time", "offline_time", "online_seconds",
		"duration_days", "created_at",
	),
	"user_subscribe": allowFields(
		"id", "user_id", "order_id", "subscribe_id", "start_time", "expire_time",
		"finished_at", "traffic", "download", "upload", "status", "note",
		"created_at", "updated_at",
	),
	"user_withdrawal": allowFields(
		"id", "user_id", "amount", "content", "status", "reason", "created_at", "updated_at",
	),
}

func allowFields(fields ...string) map[string]bool {
	allowed := make(map[string]bool, len(fields))
	for _, field := range fields {
		allowed[field] = true
	}
	return allowed
}

// StoreAdapter 将 *gorm.DB 适配为 StoreClient 接口
type StoreAdapter struct {
	db *gorm.DB
}

// NewStoreAdapter 创建数据库 Store 适配器
func NewStoreAdapter(db *gorm.DB) *StoreAdapter {
	return &StoreAdapter{db: db}
}

// Query 执行数据库查询（安全：仅白名单表 + 白名单操作）
func (a *StoreAdapter) Query(
	model string,
	operation string,
	conditions map[string]interface{},
	fields []string,
	limit, offset int32,
) ([]map[string]interface{}, int64, error) {
	if a.db == nil {
		return nil, 0, fmt.Errorf("database not available")
	}
	return a.query(model, operation, conditions, fields, limit, offset)
}

func (a *StoreAdapter) query(
	model string,
	operation string,
	conditions map[string]interface{},
	fields []string,
	limit, offset int32,
) ([]map[string]interface{}, int64, error) {
	if model == "ticket_reply" {
		return a.createTicketReply(operation, conditions)
	}

	table, ok := modelWhitelist[model]
	if !ok {
		return nil, 0, fmt.Errorf("model %q not allowed", model)
	}
	if err := validateDBFields(model, fields); err != nil {
		return nil, 0, err
	}
	if err := validateDBConditionFields(model, conditions); err != nil {
		return nil, 0, err
	}

	query := a.db.Table(table)

	// 应用查询条件
	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count %s: %w", model, err)
	}

	// 应用字段选择
	if len(fields) > 0 {
		// 安全：只选白名单字段（使用 GORM Select 会自动引用）
		query = query.Select(fields)
	}

	// 分页
	if limit > 0 {
		query = query.Limit(int(limit))
	}
	if offset > 0 {
		query = query.Offset(int(offset))
	}

	switch operation {
	case "list", "find":
		// 读取操作：返回行列表
		var rows []map[string]interface{}
		if err := query.Find(&rows).Error; err != nil {
			return nil, 0, fmt.Errorf("query %s: %w", model, err)
		}
		return rows, total, nil

	case "create":
		// 写入操作：插入
		if len(conditions) > 0 {
			err := query.Create(conditions).Error
			if err != nil {
				return nil, 0, fmt.Errorf("create %s: %w", model, err)
			}
			return []map[string]interface{}{{"affected": 1}}, 1, nil
		}
		return nil, 0, fmt.Errorf("create requires conditions")

	case "update":
		// 更新操作
		if len(conditions) > 0 {
			result := query.Updates(conditions)
			if result.Error != nil {
				return nil, 0, fmt.Errorf("update %s: %w", model, result.Error)
			}
			return []map[string]interface{}{{"affected": result.RowsAffected}}, result.RowsAffected, nil
		}
		return nil, 0, fmt.Errorf("update requires conditions")

	case "delete":
		// 删除操作
		if len(conditions) > 0 {
			result := query.Delete(nil)
			if result.Error != nil {
				return nil, 0, fmt.Errorf("delete %s: %w", model, result.Error)
			}
			return []map[string]interface{}{{"affected": result.RowsAffected}}, result.RowsAffected, nil
		}
		return nil, 0, fmt.Errorf("delete requires conditions (use specific conditions to avoid mass deletion)")

	default:
		return nil, 0, fmt.Errorf("unknown operation %q", operation)
	}
}

func (a *StoreAdapter) createTicketReply(operation string, conditions map[string]interface{}) ([]map[string]interface{}, int64, error) {
	if operation != "create" {
		return nil, 0, fmt.Errorf("unknown operation %q for ticket_reply", operation)
	}
	if err := validateDBConditionFields("ticket_reply", conditions); err != nil {
		return nil, 0, err
	}

	ticketID, ok := int64FromValue(conditions["ticket_id"])
	if !ok || ticketID <= 0 {
		return nil, 0, fmt.Errorf("ticket_reply requires ticket_id")
	}
	content := fmt.Sprint(conditions["content"])
	if content == "" {
		return nil, 0, fmt.Errorf("ticket_reply requires content")
	}
	from := fmt.Sprint(conditions["from"])
	if from == "" {
		from = "admin"
	}
	replyType, ok := int64FromValue(conditions["type"])
	if !ok || replyType <= 0 {
		replyType = 1
	}

	err := a.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("ticket").Where("id = ?", ticketID).Count(&count).Error; err != nil {
			return fmt.Errorf("find ticket: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("ticket %d not found", ticketID)
		}

		if err := tx.Table("ticket_follow").Create(map[string]interface{}{
			"ticket_id":  ticketID,
			"from":       from,
			"type":       uint8(replyType),
			"content":    content,
			"created_at": time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("create ticket follow: %w", err)
		}

		if err := tx.Table("ticket").Where("id = ?", ticketID).Updates(map[string]interface{}{
			"status":     2,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("update ticket status: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return []map[string]interface{}{{"affected": int64(1)}}, 1, nil
}

func validateDBFields(model string, fields []string) error {
	allowed := modelFieldWhitelist[model]
	if len(allowed) == 0 {
		return fmt.Errorf("model %q has no field whitelist", model)
	}
	for _, field := range fields {
		if !allowed[field] {
			return fmt.Errorf("field %q not allowed for model %q", field, model)
		}
	}
	return nil
}

func int64FromValue(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case uint64:
		if v > uint64(^uint64(0)>>1) {
			return 0, false
		}
		return int64(v), true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}

func validateDBConditionFields(model string, conditions map[string]interface{}) error {
	allowed := modelFieldWhitelist[model]
	if len(allowed) == 0 {
		return fmt.Errorf("model %q has no field whitelist", model)
	}
	for field := range conditions {
		if !allowed[field] {
			return fmt.Errorf("condition field %q not allowed for model %q", field, model)
		}
	}
	return nil
}
