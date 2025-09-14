package log

import (
	"context"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type FilterGiftLogLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Filter gift log
func NewFilterGiftLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilterGiftLogLogic {
	return &FilterGiftLogLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FilterGiftLogLogic) FilterGiftLog(req *types.FilterGiftLogRequest) (resp *types.FilterGiftLogResponse, err error) {
	data, total, err := l.svcCtx.LogModel.FilterSystemLog(l.ctx, &log.FilterParams{
		Page:     req.Page,
		Size:     req.Size,
		Type:     log.TypeGift.Uint8(),
		ObjectID: req.UserId,
		Data:     req.Date,
		Search:   req.Search,
	})

	if err != nil {
		l.Errorf("[FilterGiftLog] failed to filter system log: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "failed to filter system log: %v", err.Error())
	}

	var list []types.GiftLog
	for _, datum := range data {
		var content log.Gift
		err = content.Unmarshal([]byte(datum.Content))
		if err != nil {
			l.Errorf("[FilterGiftLog] failed to unmarshal content: %v", err.Error())
			continue
		}
		list = append(list, types.GiftLog{
			Type:        content.Type,
			UserId:      datum.ObjectID,
			OrderNo:     content.OrderNo,
			SubscribeId: content.SubscribeId,
			Amount:      content.Amount,
			Balance:     content.Balance,
			Remark:      content.Remark,
			Timestamp:   content.Timestamp,
		})
	}

	return &types.FilterGiftLogResponse{
		Total: total,
		List:  list,
	}, nil
}
