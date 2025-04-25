package subscribe

import (
	"context"
	"time"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type ResetUserSubscribePeriodLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetUserSubscribePeriodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetUserSubscribePeriodLogic {
	return &ResetUserSubscribePeriodLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetUserSubscribePeriodLogic) ResetUserSubscribePeriod(req *types.UserSubscribeResetPeriodRequest) (resp *types.UserSubscribeResetPeriodResponse, err error) {
	resp = &types.UserSubscribeResetPeriodResponse{}
	userInfo := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	subscribe, err := l.svcCtx.UserModel.FindOneSubscribe(l.ctx, req.UserSubscribeId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find order error: %v", err.Error())
	}
	if userInfo.Id != subscribe.UserId {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.SubscribeNotAvailable), "user not authorized,subscribe not  available")
	}

	if time.Now().After(subscribe.ExpireTime) {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.SubscribeExpired), "subscribe expired")
	}

	if subscribe.Traffic < 1 {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ExistAvailableTraffic), "Unlimited data plan.")
	}

	if (subscribe.Download + subscribe.Upload + 10240) < subscribe.Traffic {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ExistAvailableTraffic), "There is still available traffic.")
	}

	subscribe.ExpireTime = subscribe.ExpireTime.AddDate(0, -1, 0)
	err = l.svcCtx.UserModel.UpdateSubscribe(l.ctx, subscribe)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update subscribe error: %v", err.Error())
	}
	resp.Status = true
	return
}
