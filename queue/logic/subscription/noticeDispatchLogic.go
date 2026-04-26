package subscription

// V4.3 通知派发(决策 20 + 7.1 通知矩阵)。
// 上游(状态机 cron / reset / addon 等)往 Redis list `notice:queue` 投单条消息:
//   "<user_id>|<user_subscribe_id>|<template_key>"
// 本 cron 每 60s 拉一批,渲染并按渠道投递。
//
// 渲染数据源 — site_content 表中 `notice_<template_key>` 行,Go text/template 语法。
// 通用变量(全模板):SiteName / Timestamp / TodayDate
// 订阅相关额外:SubscribeName / RemainHuman / AddonHuman / DeviceName / OrderNo / AmountHuman / ExpireDate / Location
//
// 渠道:
//   - 邮件:必发(若用户绑定了邮箱)。直接复用现有 ForthwithSendEmail asynq 任务。
//   - Telegram:仅 device_reset 等用户操作类(见 notifyChannels 表)。
//   - 站内信:写 message_log 表(已存在),用户中心可拉取。

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"text/template"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hibiken/asynq"

	"github.com/perfect-panel/server/internal/model/message"
	"github.com/perfect-panel/server/internal/model/sitecontent"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/queue/types"
)

const noticeQueueKey = "notice:queue"
const dispatchBatch = 200 // 每次最多拉 200 条,够 5 分钟产能

// 渠道开关 — 决策 20 等的"仅邮件" / "邮件+TG+站内信" 区分。
type channels struct {
	email    bool
	telegram bool
	inMail   bool // 站内信
}

func channelsFor(templateKey string) channels {
	switch templateKey {
	case "device_reset":
		return channels{email: true, telegram: true, inMail: true}
	case "throttle_started", "throttle_12h", "cutoff", "traffic_90", "traffic_restored",
		"payment_success", "expire_3d", "expire_1d", "admin_login_remote":
		return channels{email: true} // 决策 20:仅邮件
	default:
		return channels{email: true}
	}
}

// 站内信会被默认开启,继续保留 inMail 开关。
type NoticeDispatchLogic struct {
	svc *svc.ServiceContext
}

func NewNoticeDispatchLogic(svc *svc.ServiceContext) *NoticeDispatchLogic {
	return &NoticeDispatchLogic{svc: svc}
}

func (l *NoticeDispatchLogic) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	for i := 0; i < dispatchBatch; i++ {
		raw, err := l.svc.Redis.RPop(ctx, noticeQueueKey).Result()
		if err != nil {
			break // empty list or error
		}
		l.dispatchOne(ctx, raw)
	}
	return nil
}

func (l *NoticeDispatchLogic) dispatchOne(ctx context.Context, raw string) {
	parts := strings.SplitN(raw, "|", 3)
	if len(parts) != 3 {
		logger.WithContext(ctx).Error("[NoticeDispatch] malformed entry", logger.Field("raw", raw))
		return
	}
	userId, _ := strconv.ParseInt(parts[0], 10, 64)
	userSubId, _ := strconv.ParseInt(parts[1], 10, 64)
	tplKey := parts[2]

	u, err := l.svc.UserModel.FindOne(ctx, userId)
	if err != nil || u.Id == 0 {
		return
	}
	tplRow, err := l.svc.SiteContentModel.GetWithFallback(ctx, "notice_"+tplKey, sitecontent.DefaultLang)
	if err != nil {
		logger.WithContext(ctx).Error("[NoticeDispatch] template missing",
			logger.Field("key", "notice_"+tplKey), logger.Field("error", err.Error()))
		return
	}

	data := l.buildVars(ctx, u, userSubId)
	subject := renderInline(tplRow.Title, data)
	body := renderInline(tplRow.Body, data)

	ch := channelsFor(tplKey)

	// 邮件 — 复用现有 SendEmailLogic + EmailTypeNotice 直发
	if ch.email {
		if email := l.userEmail(ctx, u); email != "" {
			payload := types.SendEmailPayload{
				Type:    types.EmailTypeNotice,
				Email:   email,
				Subject: subject,
				Content: map[string]interface{}{"Body": body},
			}
			val, _ := json.Marshal(payload)
			t := asynq.NewTask(types.ForthwithSendEmail, val, asynq.MaxRetry(3))
			if _, err := l.svc.Queue.EnqueueContext(ctx, t); err != nil {
				logger.WithContext(ctx).Error("[NoticeDispatch] enqueue email failed",
					logger.Field("error", err.Error()), logger.Field("user_id", u.Id))
			}
		}
	}
	// Telegram — 用户绑定时发
	if ch.telegram && l.svc.TelegramBot != nil {
		if chatId := l.userTelegramChat(ctx, u); chatId != 0 {
			msg := tgbotapi.NewMessage(chatId, subject+"\n\n"+body)
			if _, err := l.svc.TelegramBot.Send(msg); err != nil {
				logger.WithContext(ctx).Error("[NoticeDispatch] tg send failed",
					logger.Field("error", err.Error()), logger.Field("chat_id", chatId))
			}
		}
	}
	// 站内信 — V4.3 P1:写到 user_message 表,用户中心 GET /v1/portal/messages 拉取
	if ch.inMail {
		link := ""
		if userSubId > 0 {
			link = "/device-billing"
		}
		_ = l.svc.MessageModel.Insert(ctx, &message.UserMessage{
			UserId:   u.Id,
			Category: tplKey,
			Title:    subject,
			Body:     body,
			Link:     link,
		})
	}
	_ = userSubId
}

func (l *NoticeDispatchLogic) buildVars(ctx context.Context, u *user.User, userSubId int64) map[string]interface{} {
	now := time.Now()
	vars := map[string]interface{}{
		"SiteName":  l.svc.Config.Site.SiteName,
		"Timestamp": now.Format("2006-01-02 15:04:05"),
		"TodayDate": now.Format("2006-01-02"),
		"UserId":    u.Id,
	}
	if userSubId > 0 {
		sub, err := l.svc.UserModel.FindOneSubscribe(ctx, userSubId)
		if err == nil {
			vars["SubscribeId"] = sub.SubscribeId
			plan, _ := l.svc.SubscribeModel.FindOne(ctx, sub.SubscribeId)
			if plan != nil {
				vars["SubscribeName"] = plan.Name
			}
			quota := sub.Traffic + sub.TrafficAddon
			used := sub.Download + sub.Upload
			remain := quota - used
			if remain < 0 {
				remain = 0
			}
			vars["RemainHuman"] = humanReadableSize(remain)
			vars["UsedHuman"] = humanReadableSize(used)
			vars["QuotaHuman"] = humanReadableSize(quota)
			vars["ExpireDate"] = sub.ExpireTime.Format("2006-01-02")
		}
	}
	return vars
}

func renderInline(tpl string, data map[string]interface{}) string {
	t, err := template.New("notice").Parse(tpl)
	if err != nil {
		return tpl
	}
	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		return tpl
	}
	return sb.String()
}

// userEmail — 取邮箱认证方式;用户中心已保证只允许 1 个 email 主认证。
func (l *NoticeDispatchLogic) userEmail(ctx context.Context, u *user.User) string {
	auth, err := l.svc.UserModel.FindUserAuthMethodByUserId(ctx, "email", u.Id)
	if err != nil || auth == nil || !auth.Verified {
		return ""
	}
	return auth.AuthIdentifier
}

// userTelegramChat — 取已绑定的 telegram chat_id(若有)。
func (l *NoticeDispatchLogic) userTelegramChat(ctx context.Context, u *user.User) int64 {
	auth, err := l.svc.UserModel.FindUserAuthMethodByUserId(ctx, "telegram", u.Id)
	if err != nil || auth == nil || auth.AuthIdentifier == "" {
		return 0
	}
	v, _ := strconv.ParseInt(auth.AuthIdentifier, 10, 64)
	return v
}

// humanReadableSize 简易字节展示。10 GiB / 500 MiB / 24 KiB 这种粒度足够通知文案。
func humanReadableSize(bytes int64) string {
	const (
		_   = iota
		kib = 1 << (10 * iota)
		mib
		gib
		tib
	)
	switch {
	case bytes >= tib:
		return strconv.FormatFloat(float64(bytes)/tib, 'f', 2, 64) + " TiB"
	case bytes >= gib:
		return strconv.FormatFloat(float64(bytes)/gib, 'f', 2, 64) + " GiB"
	case bytes >= mib:
		return strconv.FormatFloat(float64(bytes)/mib, 'f', 2, 64) + " MiB"
	case bytes >= kib:
		return strconv.FormatFloat(float64(bytes)/kib, 'f', 2, 64) + " KiB"
	default:
		return strconv.FormatInt(bytes, 10) + " B"
	}
}
