package user

import (
	"context"
	"time"

	"github.com/perfect-panel/ppanel-server/pkg/constant"

	"github.com/google/uuid"
	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/uuidx"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type ResetUserSubscribeTokenLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewResetUserSubscribeTokenLogic Reset User Subscribe Token
func NewResetUserSubscribeTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetUserSubscribeTokenLogic {
	return &ResetUserSubscribeTokenLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetUserSubscribeTokenLogic) ResetUserSubscribeToken(req *types.ResetUserSubscribeTokenRequest) error {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	userSub, err := l.svcCtx.UserModel.FindOneUserSubscribe(l.ctx, req.UserSubscribeId)
	if err != nil {
		l.Errorw("FindOneUserSubscribe failed:", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FindOneUserSubscribe failed: %v", err.Error())
	}
	if userSub.UserId != u.Id {
		l.Errorw("UserSubscribeId does not belong to the current user")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "UserSubscribeId does not belong to the current user")
	}
	// find order
	orderDetails, err := l.svcCtx.OrderModel.FindOneDetails(l.ctx, userSub.OrderId)
	if err != nil {
		l.Errorw("FindOneDetails failed:", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FindOneDetails failed: %v", err.Error())
	}
	userSub.Token = uuidx.SubscribeToken(orderDetails.OrderNo + time.Now().Format("20060102150405.000"))
	userSub.UUID = uuid.New().String()
	var newSub user.Subscribe
	tool.DeepCopy(&newSub, userSub)

	err = l.svcCtx.UserModel.UpdateSubscribe(l.ctx, &newSub)
	if err != nil {
		l.Errorw("UpdateSubscribe failed:", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "UpdateSubscribe failed: %v", err.Error())
	}
	return nil
}
