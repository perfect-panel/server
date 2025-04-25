package subscribe

import (
	"context"

	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type QueryUserAlreadySubscribeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get  Already subscribed to package
func NewQueryUserAlreadySubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserAlreadySubscribeLogic {
	return &QueryUserAlreadySubscribeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryUserAlreadySubscribeLogic) QueryUserAlreadySubscribe() (resp *types.QueryUserSubscribeResp, err error) {
	resp = &types.QueryUserSubscribeResp{
		Data: make([]types.UserSubscribeData, 0),
	}
	userInfo := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	var orderIds []int64
	var subscribes []user.Subscribe
	err = l.svcCtx.OrderModel.Transaction(context.Background(), func(tx *gorm.DB) error {
		if err := tx.Model(&order.Order{}).Where("user_id = ? AND status in ?", userInfo.Id, []int64{2, 5}).Select("id").Find(&orderIds).Error; err != nil {
			return err
		}
		if len(orderIds) == 0 {
			return nil
		}
		return tx.Model(&user.Subscribe{}).Where("user_id = ? AND order_id in ?", userInfo.Id, orderIds).Order("created_at desc").Find(&subscribes).Error
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find order error: %v", err.Error())
	}
	if len(subscribes) == 0 {
		return
	}

	userAlreadySubscribe := make(map[int64]int64)
	for _, subscribe := range subscribes {
		userAlreadySubscribe[subscribe.SubscribeId] = subscribe.Id
	}

	for k, v := range userAlreadySubscribe {
		resp.Data = append(resp.Data, types.UserSubscribeData{
			SubscribeId:     k,
			UserSubscribeId: v,
		})
	}
	return
}
