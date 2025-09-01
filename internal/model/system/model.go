package system

import (
	"context"

	"github.com/perfect-panel/server/internal/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type customSystemLogicModel interface {
	GetSmsConfig(ctx context.Context) ([]*System, error)
	GetSiteConfig(ctx context.Context) ([]*System, error)
	GetSubscribeConfig(ctx context.Context) ([]*System, error)
	GetRegisterConfig(ctx context.Context) ([]*System, error)
	GetVerifyConfig(ctx context.Context) ([]*System, error)
	GetNodeConfig(ctx context.Context) ([]*System, error)
	GetInviteConfig(ctx context.Context) ([]*System, error)
	GetTosConfig(ctx context.Context) ([]*System, error)
	GetCurrencyConfig(ctx context.Context) ([]*System, error)
	GetVerifyCodeConfig(ctx context.Context) ([]*System, error)
	GetLogConfig(ctx context.Context) ([]*System, error)
	UpdateNodeMultiplierConfig(ctx context.Context, config string) error
	FindNodeMultiplierConfig(ctx context.Context) (*System, error)
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customSystemModel{
		defaultSystemModel: newSystemModel(conn, c),
	}
}

// GetSmsConfig returns the sms config.
func (m *customSystemModel) GetSmsConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.SmsConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "sms").Find(v).Error
	})
	return configs, err
}

// GetSiteConfig returns the site config.
func (m *customSystemModel) GetSiteConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.SiteConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "site").Find(v).Error
	})
	return configs, err
}

// GetEmailConfig returns the email config.
func (m *customSystemModel) GetEmailConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.EmailSmtpConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "email").Find(v).Error
	})
	return configs, err
}

// GetSubscribeConfig returns the subscribe config.
func (m *customSystemModel) GetSubscribeConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.SubscribeConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "subscribe").Find(v).Error
	})
	return configs, err
}

// GetRegisterConfig returns the register config.
func (m *customSystemModel) GetRegisterConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.RegisterConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "register").Find(v).Error
	})
	return configs, err
}

// GetVerifyConfig returns the verify config.
func (m *customSystemModel) GetVerifyConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.VerifyConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "verify").Find(v).Error
	})
	return configs, err
}

// GetNodeConfig returns the server config.
func (m *customSystemModel) GetNodeConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.NodeConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "server").Find(v).Error
	})
	return configs, err
}

// GetInviteConfig returns the invite config.
func (m *customSystemModel) GetInviteConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.InviteConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "invite").Find(v).Error
	})
	return configs, err
}

// GetTelegramConfig returns the telegram config.
func (m *customSystemModel) GetTelegramConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.TelegramConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "telegram").Find(v).Error
	})
	return configs, err
}

// GetTosConfig returns the tos config.
func (m *customSystemModel) GetTosConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.TosConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "tos").Find(v).Error
	})
	return configs, err
}

// GetCurrencyConfig returns the currency config.
func (m *customSystemModel) GetCurrencyConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.CurrencyConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "currency").Find(v).Error
	})
	return configs, err
}

func (m *customSystemModel) UpdateNodeMultiplierConfig(ctx context.Context, config string) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return conn.Model(&System{}).Where("`category` = ? AND `key` = ?", "server", "NodeMultiplierConfig").Update("value", config).Error
	})
}

func (m *customSystemModel) FindNodeMultiplierConfig(ctx context.Context) (*System, error) {
	var data System
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ? AND `key` = ?", "server", "NodeMultiplierConfig").Find(v).Error
	})
	return &data, err
}

// GetVerifyCodeConfig returns the verify code config.

func (m *customSystemModel) GetVerifyCodeConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryCtx(ctx, &configs, config.VerifyCodeConfigKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "verify_code").Find(v).Error
	})
	return configs, err
}

// GetLogConfig returns the log config.
func (m *customSystemModel) GetLogConfig(ctx context.Context) ([]*System, error) {
	var configs []*System
	err := m.QueryNoCacheCtx(ctx, &configs, func(conn *gorm.DB, v interface{}) error {
		return conn.Where("`category` = ?", "log").Find(v).Error
	})
	return configs, err
}
