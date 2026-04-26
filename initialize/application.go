package initialize

// V4.3 决策 25:确保 11 款官方客户端在 subscribe_application 表中存在,
// 每个挂载到 site_content 里对应的多语言教程 key。
//
// 行为:
//   1. 软删除遗留的 name='Default' 行(V4.3 决策:不再保留兜底假客户端)
//   2. 对照 v43Clients 列表:按 name 查表,缺哪个补哪个
//   3. 已存在的客户端,如果 tutorial_key 为空则补上,其它字段不动
//
// 这样老站升级 V4.3 时不会清掉管理员自定义的客户端,只是补齐 V4.3 默认列表 +
// 给现有同名客户端关联教程。

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/server/internal/model/client"
	"github.com/perfect-panel/server/internal/model/sitecontent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"gorm.io/gorm"
)

type v43Client struct {
	Name         string
	UserAgent    string
	Scheme       string
	OutputFormat string
	Description  string
	TutorialKey  string
	DownloadLink client.DownloadLink
}

// V4.3 11 款 + 数据库已有的 SingBox(V4.3 没列但保留 = 12 款)。
// 见 docs/落地方案_V4.3.md SQL §3.2 (6) site_content 默认填充。
var v43Clients = []v43Client{
	{
		Name:         "v2rayN",
		UserAgent:    "v2rayN",
		Scheme:       "",
		OutputFormat: "Base64",
		Description:  "V2Ray Windows GUI client",
		TutorialKey:  sitecontent.KeyClientTutorialV2rayN,
		DownloadLink: client.DownloadLink{
			Windows: "https://github.com/2dust/v2rayN/releases/latest",
		},
	},
	{
		Name:         "Clash",
		UserAgent:    "Clash",
		Scheme:       "clash://install-config?url=",
		OutputFormat: "YAML",
		Description:  "Cross-platform proxy client",
		TutorialKey:  sitecontent.KeyClientTutorialClash,
		DownloadLink: client.DownloadLink{
			Windows: "https://github.com/Fndroid/clash_for_windows_pkg/releases",
		},
	},
	{
		Name:         "Stash",
		UserAgent:    "Stash",
		Scheme:       "stash://install-config?url=",
		OutputFormat: "YAML",
		Description:  "iOS rule-based proxy client",
		TutorialKey:  sitecontent.KeyClientTutorialStash,
		DownloadLink: client.DownloadLink{
			IOS: "https://apps.apple.com/app/stash/id1596063349",
		},
	},
	{
		Name:         "Shadowrocket",
		UserAgent:    "Shadowrocket",
		Scheme:       "shadowrocket://add/sub://",
		OutputFormat: "Base64",
		Description:  "iOS / iPadOS shadowsocks client",
		TutorialKey:  sitecontent.KeyClientTutorialShadowrocket,
		DownloadLink: client.DownloadLink{
			IOS: "https://apps.apple.com/app/shadowrocket/id932747118",
		},
	},
	{
		Name:         "Hiddify",
		UserAgent:    "Hiddify",
		Scheme:       "hiddify://install-config?url=",
		OutputFormat: "JSON",
		Description:  "Cross-platform sing-box GUI",
		TutorialKey:  sitecontent.KeyClientTutorialHiddify,
		DownloadLink: client.DownloadLink{
			Windows: "https://github.com/hiddify/hiddify-next/releases",
			Android: "https://github.com/hiddify/hiddify-next/releases",
			IOS:     "https://apps.apple.com/app/hiddify-proxy-vpn/id6596777532",
			Mac:     "https://github.com/hiddify/hiddify-next/releases",
		},
	},
	{
		Name:         "Clash Meta for Android",
		UserAgent:    "ClashMeta",
		Scheme:       "clashmeta://install-config?url=",
		OutputFormat: "YAML",
		Description:  "Clash.Meta Android client",
		TutorialKey:  sitecontent.KeyClientTutorialClashMeta,
		DownloadLink: client.DownloadLink{
			Android: "https://github.com/MetaCubeX/ClashMetaForAndroid/releases",
		},
	},
	{
		Name:         "Quantumult X",
		UserAgent:    "Quantumult",
		Scheme:       "quantumult-x:///add-resource?remote-resource=",
		OutputFormat: "CONF",
		Description:  "iOS network debugging tool",
		TutorialKey:  sitecontent.KeyClientTutorialQuantumult,
		DownloadLink: client.DownloadLink{
			IOS: "https://apps.apple.com/app/quantumult-x/id1443988620",
		},
	},
	{
		Name:         "Loon",
		UserAgent:    "Loon",
		Scheme:       "loon://import?nodelist=",
		OutputFormat: "CONF",
		Description:  "iOS network proxy & rules",
		TutorialKey:  sitecontent.KeyClientTutorialLoon,
		DownloadLink: client.DownloadLink{
			IOS: "https://apps.apple.com/app/loon/id1373567447",
		},
	},
	{
		Name:         "FlClash",
		UserAgent:    "FlClash",
		Scheme:       "",
		OutputFormat: "YAML",
		Description:  "Flutter-based Clash GUI",
		TutorialKey:  sitecontent.KeyClientTutorialFlClash,
		DownloadLink: client.DownloadLink{
			Windows: "https://github.com/chen08209/FlClash/releases",
			Android: "https://github.com/chen08209/FlClash/releases",
			Linux:   "https://github.com/chen08209/FlClash/releases",
		},
	},
	{
		Name:         "Surge",
		UserAgent:    "Surge",
		Scheme:       "surge:///install-config?url=",
		OutputFormat: "CONF",
		Description:  "Premium iOS network toolbox",
		TutorialKey:  sitecontent.KeyClientTutorialSurge,
		DownloadLink: client.DownloadLink{
			IOS: "https://apps.apple.com/app/surge-5/id1442620678",
		},
	},
	{
		Name:         "Surge for Mac",
		UserAgent:    "Surge Mac",
		Scheme:       "surge:///install-config?url=",
		OutputFormat: "CONF",
		Description:  "Premium macOS network toolbox",
		TutorialKey:  sitecontent.KeyClientTutorialSurgeMac,
		DownloadLink: client.DownloadLink{
			Mac: "https://nssurge.com/",
		},
	},
}

func Application(svcCtx *svc.ServiceContext) {
	logger.Debug("[Init Application] V4.3 client seed")
	ctx := context.Background()

	db := svcCtx.DB
	// 1) 删除遗留 'Default' 客户端(V4.3 决策:不再保留兜底假客户端)
	if err := db.WithContext(ctx).
		Where("name = ?", "Default").
		Delete(&client.SubscribeApplication{}).Error; err != nil {
		logger.Errorf("[Init Application] failed to clean legacy Default: %v", err)
	}

	// 2) 取现有客户端,按 name 建立索引(规避 unique 冲突)
	existing, err := svcCtx.ClientModel.List(ctx)
	if err != nil {
		logger.Errorf("[Init Application] list failed: %v", err)
		return
	}
	byName := make(map[string]*client.SubscribeApplication, len(existing))
	for _, e := range existing {
		byName[e.Name] = e
	}

	// 3) 缺哪个补哪个;已存在但 tutorial_key 为空的回填
	for i := range v43Clients {
		c := &v43Clients[i]
		if cur, ok := byName[c.Name]; ok {
			if cur.TutorialKey == "" && c.TutorialKey != "" {
				cur.TutorialKey = c.TutorialKey
				if err := db.WithContext(ctx).
					Model(&client.SubscribeApplication{}).
					Where("id = ?", cur.Id).
					Update("tutorial_key", c.TutorialKey).Error; err != nil {
					logger.Errorf("[Init Application] backfill tutorial_key for %s: %v", c.Name, err)
				}
			}
			continue
		}
		linkData, _ := json.Marshal(c.DownloadLink)
		row := &client.SubscribeApplication{
			Name:              c.Name,
			Description:       c.Description,
			Scheme:            c.Scheme,
			UserAgent:         c.UserAgent,
			OutputFormat:      c.OutputFormat,
			DownloadLink:      string(linkData),
			TutorialKey:       c.TutorialKey,
			IsDefault:         false,
			Enabled:           true, // 默认启用,管理员可后续关闭
			SubscribeTemplate: "",
		}
		if err := svcCtx.ClientModel.Insert(ctx, row); err != nil {
			// 如果是并发种子或唯一约束失败,降级为日志
			logger.Errorf("[Init Application] insert %s: %v", c.Name, err)
			continue
		}
		// 顺手保证 site_content 里有对应教程占位行(V4.3 SQL 已 seed,
		// 这里再 Upsert 是幂等兜底,防止数据库被外部清过)
		_ = svcCtx.SiteContentModel.Upsert(ctx, &sitecontent.SiteContent{
			ContentKey:  c.TutorialKey,
			ContentLang: sitecontent.DefaultLang,
			Title:       c.Name + " 使用教程",
			Body:        "<管理员后续编辑>",
		})
	}
}

// 引入未使用的 gorm 类型避免静态 import 警告(后续若加 hook 需要)
var _ = (*gorm.DB)(nil)
