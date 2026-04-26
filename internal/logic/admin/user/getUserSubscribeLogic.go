package user

import (
	"context"
	"time"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetUserSubscribeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user subcribe
func NewGetUserSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSubscribeLogic {
	return &GetUserSubscribeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSubscribeLogic) GetUserSubscribe(req *types.GetUserSubscribeListRequest) (resp *types.GetUserSubscribeListResponse, err error) {
	data, err := l.svcCtx.UserModel.QueryUserSubscribe(l.ctx, req.UserId, 0, 1, 2, 3, 4)
	if err != nil {
		l.Errorw("[GetUserSubscribeLogs] Get User Subscribe Error:", logger.Field("err", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Get User Subscribe Error")
	}

	resp = &types.GetUserSubscribeListResponse{
		List:  make([]types.UserSubscribe, 0),
		Total: int64(len(data)),
	}

	for _, item := range data {
		var sub types.UserSubscribe
		tool.DeepCopy(&sub, item)
		sub.Short, _ = tool.FixedUniqueString(item.Token, 8, "")
		sub.ResetTime = calculateNextResetTime(&sub)
		resp.List = append(resp.List, sub)
	}
	return
}

// calculateNextResetTime — 与 public/user/queryUserSubscribeLogic.go 同语义,根据
// 套餐 reset_cycle 算下次自动清零的毫秒时间戳。0 = 不重置。
func calculateNextResetTime(sub *types.UserSubscribe) int64 {
	if sub.Subscribe.ResetCycle == 0 {
		return 0
	}
	resetTime := time.UnixMilli(sub.ExpireTime)
	now := time.Now()
	switch sub.Subscribe.ResetCycle {
	case 1:
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).UnixMilli()
	case 2:
		if resetTime.Day() > now.Day() {
			return time.Date(now.Year(), now.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
		}
		return time.Date(now.Year(), now.Month()+1, resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	case 3:
		targetTime := time.Date(now.Year(), resetTime.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location())
		if targetTime.Before(now) {
			targetTime = time.Date(now.Year()+1, resetTime.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location())
		}
		return targetTime.UnixMilli()
	default:
		return 0
	}
}
