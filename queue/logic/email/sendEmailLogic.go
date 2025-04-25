package emailLogic

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/ppanel-server/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/ppanel-server/internal/model/log"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/email"
	"github.com/perfect-panel/ppanel-server/queue/types"
)

type SendEmailLogic struct {
	svcCtx *svc.ServiceContext
}

func NewSendEmailLogic(svcCtx *svc.ServiceContext) *SendEmailLogic {
	return &SendEmailLogic{
		svcCtx: svcCtx,
	}
}
func (l *SendEmailLogic) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload types.SendEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logger.WithContext(ctx).Error("[SendEmailLogic] Unmarshal payload failed",
			logger.Field("error", err.Error()),
			logger.Field("payload", task.Payload()),
		)
		return nil
	}
	messageLog := log.MessageLog{
		Type:     log.Email.String(),
		Platform: l.svcCtx.Config.Email.Platform,
		To:       payload.Email,
		Subject:  payload.Subject,
		Content:  payload.Content,
	}
	sender, err := email.NewSender(l.svcCtx.Config.Email.Platform, l.svcCtx.Config.Email.PlatformConfig, l.svcCtx.Config.Site.SiteName)
	if err != nil {
		logger.WithContext(ctx).Error("[SendEmailLogic] NewSender failed", logger.Field("error", err.Error()))
		return nil
	}
	err = sender.Send([]string{payload.Email}, payload.Subject, payload.Content)
	if err != nil {
		logger.WithContext(ctx).Error("[SendEmailLogic] Send email failed", logger.Field("error", err.Error()))
		return nil
	}
	messageLog.Status = 1
	if err = l.svcCtx.LogModel.InsertMessageLog(ctx, &messageLog); err != nil {
		logger.WithContext(ctx).Error("[SendEmailLogic] InsertMessageLog failed",
			logger.Field("error", err.Error()),
			logger.Field("messageLog", messageLog),
		)
	}
	logger.WithContext(ctx).Info("[SendEmailLogic] Send email", logger.Field("email", payload.Email), logger.Field("content", payload.Content))
	return nil
}
