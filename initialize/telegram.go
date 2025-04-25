package initialize

import (
	"context"
	"fmt"

	"github.com/perfect-panel/server/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/logic/telegram"
	"github.com/perfect-panel/server/internal/model/auth"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/tool"
)

func Telegram(svc *svc.ServiceContext) {

	method, err := svc.AuthModel.FindOneByMethod(context.Background(), "telegram")
	if err != nil {
		logger.Errorf("[Init Telegram Config] Get Telegram Config Error: %s", err.Error())
		return
	}
	var tg config.Telegram

	tgConfig := new(auth.TelegramAuthConfig)
	if err = tgConfig.Unmarshal(method.Config); err != nil {
		logger.Errorf("[Init Telegram Config] Unmarshal Telegram Config Error: %s", err.Error())
		return
	}

	if tgConfig.BotToken == "" {
		logger.Debug("[Init Telegram Config] Telegram Token is empty")
		return
	}

	bot, err := tgbotapi.NewBotAPI(tg.BotToken)
	if err != nil {
		logger.Error("[Init Telegram Config] New Bot API Error: ", logger.Field("error", err.Error()))
		return
	}

	if tgConfig.WebHookDomain == "" || svc.Config.Debug {
		// set Long Polling mode
		updateConfig := tgbotapi.NewUpdate(0)
		updateConfig.Timeout = 60
		updates := bot.GetUpdatesChan(updateConfig)
		go func() {
			for update := range updates {
				if update.Message != nil {
					ctx := context.Background()
					l := telegram.NewTelegramLogic(ctx, svc)
					l.TelegramLogic(&update)
				}
			}
		}()
	} else {
		wh, err := tgbotapi.NewWebhook(fmt.Sprintf("%s/v1/telegram/webhook?secret=%s", tgConfig.WebHookDomain, tool.Md5Encode(tgConfig.BotToken, false)))
		if err != nil {
			logger.Errorf("[Init Telegram Config] New Webhook Error: %s", err.Error())
			return
		}
		_, err = bot.Request(wh)
		if err != nil {
			logger.Errorf("[Init Telegram Config] Request Webhook Error: %s", err.Error())
			return
		}
	}

	user, err := bot.GetMe()
	if err != nil {
		logger.Error("[Init Telegram Config] Get Bot Info Error: ", logger.Field("error", err.Error()))
		return
	}
	svc.Config.Telegram.BotID = user.ID
	svc.Config.Telegram.BotName = user.UserName
	svc.Config.Telegram.EnableNotify = tg.EnableNotify
	svc.Config.Telegram.WebHookDomain = tg.WebHookDomain
	svc.TelegramBot = bot

	logger.Info("[Init Telegram Config] Webhook set success")
}
