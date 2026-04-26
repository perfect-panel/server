package sitecontent

// V4.3 site_content — 用户协议 + 11 款客户端教程(决策 19 / 25)。
// 内容由管理端 CMS 维护,前端按 (key, lang) 取。

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// 内置 key:写死保证 enum 一致;新加请同步前端字典。
const (
	KeyTermsOfUse                = "terms_of_use"
	KeyClientTutorialV2rayN      = "client_tutorial_v2rayn"
	KeyClientTutorialClash       = "client_tutorial_clash"
	KeyClientTutorialStash       = "client_tutorial_stash"
	KeyClientTutorialShadowrocket = "client_tutorial_shadowrocket"
	KeyClientTutorialHiddify     = "client_tutorial_hiddify"
	KeyClientTutorialClashMeta   = "client_tutorial_clashmeta"
	KeyClientTutorialQuantumult  = "client_tutorial_quantumult"
	KeyClientTutorialLoon        = "client_tutorial_loon"
	KeyClientTutorialFlClash     = "client_tutorial_flclash"
	KeyClientTutorialSurge       = "client_tutorial_surge"
	KeyClientTutorialSurgeMac    = "client_tutorial_surge_mac"

	DefaultLang = "zh-CN"
)

type SiteContent struct {
	Id          int64     `gorm:"primaryKey"`
	ContentKey  string    `gorm:"column:content_key;type:varchar(64);not null;default:'';uniqueIndex:uni_site_content_key_lang,priority:1"`
	ContentLang string    `gorm:"column:content_lang;type:varchar(8);not null;default:'zh-CN';uniqueIndex:uni_site_content_key_lang,priority:2"`
	Title       string    `gorm:"column:title;type:varchar(255);not null;default:''"`
	Body        string    `gorm:"column:body;type:mediumtext"`
	// V4.4 #45:管理员手工 bump,用户已接受版本与此不一致时需重新接受。
	// 同 key 不同 lang 的 version 可以不一致(各语言独立审校),前端按当前 lang 取。
	Version   string    `gorm:"column:version;type:varchar(32);not null;default:'1'"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}

func (*SiteContent) TableName() string {
	return "site_content"
}

type Model interface {
	Get(ctx context.Context, key, lang string) (*SiteContent, error)
	GetWithFallback(ctx context.Context, key, lang string) (*SiteContent, error)
	Upsert(ctx context.Context, data *SiteContent) error
	List(ctx context.Context, lang string) ([]*SiteContent, error)
	ListByPrefix(ctx context.Context, prefix, lang string) ([]*SiteContent, error)
}

type defaultModel struct {
	db *gorm.DB
}

func NewModel(db *gorm.DB) Model {
	return &defaultModel{db: db}
}

func (m *defaultModel) Get(ctx context.Context, key, lang string) (*SiteContent, error) {
	var data SiteContent
	err := m.db.WithContext(ctx).
		Where("content_key = ? AND content_lang = ?", key, lang).
		First(&data).Error
	return &data, err
}

// GetWithFallback 优先取目标语言,缺失则回退 zh-CN(默认占位)。
func (m *defaultModel) GetWithFallback(ctx context.Context, key, lang string) (*SiteContent, error) {
	data, err := m.Get(ctx, key, lang)
	if err == nil {
		return data, nil
	}
	if lang == DefaultLang {
		return data, err
	}
	return m.Get(ctx, key, DefaultLang)
}

// Upsert: ON DUPLICATE KEY UPDATE on (content_key, content_lang)。
// data.Version 为空字符串则保留旧版本(不 bump)。
func (m *defaultModel) Upsert(ctx context.Context, data *SiteContent) error {
	assign := map[string]interface{}{
		"title": data.Title,
		"body":  data.Body,
	}
	if data.Version != "" {
		assign["version"] = data.Version
	}
	return m.db.WithContext(ctx).
		Where("content_key = ? AND content_lang = ?", data.ContentKey, data.ContentLang).
		Assign(assign).
		FirstOrCreate(data).Error
}

func (m *defaultModel) List(ctx context.Context, lang string) ([]*SiteContent, error) {
	var list []*SiteContent
	conn := m.db.WithContext(ctx).Model(&SiteContent{})
	if lang != "" {
		conn = conn.Where("content_lang = ?", lang)
	}
	err := conn.Order("content_key ASC").Find(&list).Error
	return list, err
}

func (m *defaultModel) ListByPrefix(ctx context.Context, prefix, lang string) ([]*SiteContent, error) {
	var list []*SiteContent
	conn := m.db.WithContext(ctx).Model(&SiteContent{}).
		Where("content_key LIKE ?", prefix+"%")
	if lang != "" {
		conn = conn.Where("content_lang = ?", lang)
	}
	err := conn.Order("content_key ASC").Find(&list).Error
	return list, err
}
