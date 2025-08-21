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

type GetMessageLogListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetMessageLogListLogic Get message log list
func NewGetMessageLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessageLogListLogic {
	return &GetMessageLogListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMessageLogListLogic) GetMessageLogList(req *types.GetMessageLogListRequest) (resp *types.GetMessageLogListResponse, err error) {

	data, total, err := l.svcCtx.LogModel.FilterSystemLog(l.ctx, &log.FilterParams{
		Page:   req.Page,
		Size:   req.Size,
		Type:   req.Type,
		Search: req.Search,
	})

	if err != nil {
		l.Errorf("[GetMessageLogList] failed to filter system log: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "failed to filter system log: %v", err.Error())
	}

	var list []types.MessageLog

	for _, datum := range data {
		var content log.Message
		err = content.Unmarshal([]byte(datum.Content))
		if err != nil {
			l.Errorf("[GetMessageLogList] failed to unmarshal content: %v", err.Error())
			continue
		}
		list = append(list, types.MessageLog{
			Id:        datum.Id,
			Type:      datum.Type,
			Platform:  content.Platform,
			To:        content.To,
			Subject:   content.Subject,
			Content:   content.Content,
			Status:    content.Status,
			CreatedAt: datum.CreatedAt.UnixMilli(),
		})
	}

	return &types.GetMessageLogListResponse{
		Total: total,
		List:  list,
	}, nil
}
