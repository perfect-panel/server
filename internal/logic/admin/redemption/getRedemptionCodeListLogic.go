package redemption

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetRedemptionCodeListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get redemption code list
func NewGetRedemptionCodeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRedemptionCodeListLogic {
	return &GetRedemptionCodeListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRedemptionCodeListLogic) GetRedemptionCodeList(req *types.GetRedemptionCodeListRequest) (resp *types.GetRedemptionCodeListResponse, err error) {
	total, list, err := l.svcCtx.RedemptionCodeModel.QueryRedemptionCodeListByPage(
		l.ctx,
		int(req.Page),
		int(req.Size),
		req.SubscribePlan,
		req.UnitTime,
		req.Code,
	)
	if err != nil {
		l.Errorw("[GetRedemptionCodeList] Database Error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get redemption code list error: %v", err.Error())
	}

	var redemptionCodes []types.RedemptionCode
	for _, item := range list {
		redemptionCodes = append(redemptionCodes, types.RedemptionCode{
			Id:            item.Id,
			Code:          item.Code,
			TotalCount:    item.TotalCount,
			UsedCount:     item.UsedCount,
			SubscribePlan: item.SubscribePlan,
			UnitTime:      item.UnitTime,
			Quantity:      item.Quantity,
			CreatedAt:     item.CreatedAt.Unix(),
			UpdatedAt:     item.UpdatedAt.Unix(),
		})
	}

	return &types.GetRedemptionCodeListResponse{
		Total: total,
		List:  redemptionCodes,
	}, nil
}
