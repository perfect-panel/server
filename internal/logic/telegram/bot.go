package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/perfect-panel/server/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/model/auth"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

func GetTelegramConfig(ctx context.Context, svcCtx *svc.ServiceContext) (*types.TelegramConfig, error) {

	data, err := svcCtx.AuthModel.FindOneByMethod(ctx, "telegram")
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get Telegram config failed: %v", err.Error())
	}
	var telegramConfig auth.TelegramAuthConfig
	err = json.Unmarshal([]byte(data.Config), &telegramConfig)
	if err != nil {
		logger.WithContext(ctx).Error("unmarshal telegram config failed", logger.Field("error", err.Error()))
		return nil, err
	}

	return &types.TelegramConfig{
		TelegramBotToken:      telegramConfig.BotToken,
		TelegramNotify:        *data.Enabled,
		TelegramWebHookDomain: telegramConfig.WebHookDomain,
	}, nil
}

func ApiLink(ctx *gin.Context, svcCtx *svc.ServiceContext, method string) string {
	cfg, _ := GetTelegramConfig(ctx, svcCtx)
	return "https://api.telegram.org/bot" + cfg.TelegramBotToken + "/" + method
}

func SendUserMessage(ctx *gin.Context, svcCtx *svc.ServiceContext, u user.User, text string, parseMode string) {
	req, _ := http.NewRequest("GET", ApiLink(ctx, svcCtx, "sendMessage"), nil)
	q := req.URL.Query()

	userTelegramChatId, ok := findTelegram(&u)
	if !ok {
		return
	}
	q.Add("chat_id", strconv.FormatInt(userTelegramChatId, 10))
	if parseMode == "markdown" {
		text = strings.ReplaceAll(text, "_", "\\_")
	}
	q.Add("text", text)
	q.Add("parse_mode", parseMode)
	req.URL.RawQuery = q.Encode()
	_, _ = http.DefaultClient.Do(req)

}

func SendAdminMessage(ctx *gin.Context, svcCtx *svc.ServiceContext, text string, parseMode string) {
	var adminTelegram []int64
	f := false
	adminTelegramJson, err := svcCtx.Redis.Get(ctx, "adminTelegram").Result()
	if err == nil {
		err = json.Unmarshal([]byte(adminTelegramJson), &adminTelegram)
		if err == nil {
			f = true
		}
	}
	if !f {
		svcCtx.DB.Model(&user.User{}).Where("is_admin = true").Pluck("telegram", &adminTelegram)
		val, _ := json.Marshal(adminTelegram)
		_ = svcCtx.Redis.Set(ctx, "TelegramConfig", string(val), time.Duration(3600)*time.Second).Err()
	}
	req, _ := http.NewRequest("GET", ApiLink(ctx, svcCtx, "sendMessage"), nil)
	q := req.URL.Query()
	if parseMode == "markdown" {
		text = strings.ReplaceAll(text, "_", "\\_")
	}
	q.Add("text", text)
	q.Add("parse_mode", parseMode)
	for _, telegram := range adminTelegram {
		q.Add("chat_id", strconv.FormatInt(telegram, 10))
		req.URL.RawQuery = q.Encode()
		_, _ = http.DefaultClient.Do(req)
	}
}

func SetWebhook(ctx *gin.Context, svcCtx *svc.ServiceContext) error {
	configs, _ := svcCtx.SystemModel.GetSiteConfig(ctx)
	cfg := &types.SiteConfig{}
	tool.SystemConfigSliceReflectToStruct(configs, cfg)
	req, _ := http.NewRequest("GET", ApiLink(ctx, svcCtx, "setWebhook"), nil)
	q := req.URL.Query()
	q.Add("url", cfg.Host+"/telegram/webhook")
	req.URL.RawQuery = q.Encode()
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "set webhook error: %v", err)
	}
	return nil
}

func findTelegram(u *user.User) (int64, bool) {
	for _, item := range u.AuthMethods {
		if item.AuthType == "telegram" {
			// string to int64
			parseInt, err := strconv.ParseInt(item.AuthIdentifier, 10, 64)
			if err != nil {
				return 0, false
			}
			return parseInt, true
		}

	}
	return 0, false
}
