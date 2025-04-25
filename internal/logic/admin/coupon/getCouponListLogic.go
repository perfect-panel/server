package coupon

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetCouponListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get coupon list
func NewGetCouponListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCouponListLogic {
	return &GetCouponListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCouponListLogic) GetCouponList(req *types.GetCouponListRequest) (resp *types.GetCouponListResponse, err error) {
	resp = &types.GetCouponListResponse{}
	// get coupon list from db
	total, list, err := l.svcCtx.CouponModel.QueryCouponListByPage(l.ctx, int(req.Page), int(req.Size), req.Subscribe, req.Search)
	if err != nil {
		l.Errorw("[GetCouponList] Database Error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get coupon list error: %v", err.Error())
	}
	resp.Total = total
	resp.List = make([]types.Coupon, 0)
	for _, coupon := range list {
		couponInfo := types.Coupon{}
		tool.DeepCopy(&couponInfo, coupon)
		couponInfo.Subscribe = tool.StringToInt64Slice(coupon.Subscribe)
		resp.List = append(resp.List, couponInfo)
	}
	return
}
