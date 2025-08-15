package user

import (
	"context"
	"encoding/json"
	"time"

	"github.com/perfect-panel/server/pkg/constant"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryUserSubscribeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query User Subscribe
func NewQueryUserSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserSubscribeLogic {
	return &QueryUserSubscribeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryUserSubscribeLogic) QueryUserSubscribe() (resp *types.QueryUserSubscribeListResponse, err error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	data, err := l.svcCtx.UserModel.QueryUserSubscribe(l.ctx, u.Id, 0, 1, 2, 3)
	if err != nil {
		l.Errorw("[QueryUserSubscribeLogic] Query User Subscribe Error:", logger.Field("err", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Query User Subscribe Error")
	}

	resp = &types.QueryUserSubscribeListResponse{
		List:  make([]types.UserSubscribe, 0),
		Total: int64(len(data)),
	}

	for _, item := range data {
		var sub types.UserSubscribe
		tool.DeepCopy(&sub, item)

		// 解析Discount字段 避免在续订时只能续订一个月
		if item.Subscribe != nil && item.Subscribe.Discount != "" {
			var discounts []types.SubscribeDiscount
			if err := json.Unmarshal([]byte(item.Subscribe.Discount), &discounts); err == nil {
				sub.Subscribe.Discount = discounts
			}
		}

		sub.ResetTime = calculateNextResetTime(&sub)
		resp.List = append(resp.List, sub)
	}
	return
}

// 计算下次重置时间
func calculateNextResetTime(sub *types.UserSubscribe) int64 {
	resetTime := time.UnixMilli(sub.ExpireTime)
	now := time.Now()
	switch sub.Subscribe.ResetCycle {
	case 0:
		return 0
	case 1:
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).UnixMilli()
	case 2:
		if resetTime.Day() > now.Day() {
			return time.Date(now.Year(), now.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
		} else {
			return time.Date(now.Year(), now.Month()+1, resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
		}
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
