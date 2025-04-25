package document

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryDocumentListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get document list
func NewQueryDocumentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDocumentListLogic {
	return &QueryDocumentListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDocumentListLogic) QueryDocumentList() (resp *types.QueryDocumentListResponse, err error) {
	total, data, err := l.svcCtx.DocumentModel.GetDocumentListByAll(l.ctx)
	if err != nil {
		l.Errorw("[QueryDocumentList] error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "QueryDocumentList error: %v", err.Error())
	}
	resp = &types.QueryDocumentListResponse{
		Total: total,
		List:  make([]types.Document, 0),
	}
	for _, item := range data {
		resp.List = append(resp.List, types.Document{
			Id:        item.Id,
			Title:     item.Title,
			Tags:      tool.StringMergeAndRemoveDuplicates(item.Tags),
			UpdatedAt: item.UpdatedAt.UnixMilli(),
		})
	}
	return
}
