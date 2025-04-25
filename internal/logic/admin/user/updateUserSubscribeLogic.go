package user

import (
	"context"
	"time"

	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdateUserSubscribeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateUserSubscribeLogic Update user subscribe
func NewUpdateUserSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserSubscribeLogic {
	return &UpdateUserSubscribeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserSubscribeLogic) UpdateUserSubscribe(req *types.UpdateUserSubscribeRequest) error {
	userSub, err := l.svcCtx.UserModel.FindOneUserSubscribe(l.ctx, req.UserSubscribeId)
	if err != nil {
		l.Errorw("FindOneUserSubscribe failed:", logger.Field("error", err.Error()), logger.Field("userSubscribeId", req.UserSubscribeId))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FindOneUserSubscribe failed: %v", err.Error())
	}
	err = l.svcCtx.UserModel.UpdateSubscribe(l.ctx, &user.Subscribe{
		Id:          req.UserSubscribeId,
		UserId:      userSub.UserId,
		OrderId:     userSub.OrderId,
		SubscribeId: req.SubscribeId,
		StartTime:   userSub.StartTime,
		ExpireTime:  time.UnixMilli(req.ExpiredAt),
		Traffic:     req.Traffic,
		Download:    req.Download,
		Upload:      req.Upload,
		Token:       userSub.Token,
		UUID:        userSub.UUID,
		Status:      userSub.Status,
	})

	if err != nil {
		l.Errorw("UpdateSubscribe failed:", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "UpdateSubscribe failed: %v", err.Error())
	}

	return nil
}
