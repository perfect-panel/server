package common

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetClientLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get Client
func NewGetClientLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClientLogic {
	return &GetClientLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClientLogic) GetClient() (resp *types.GetSubscribeClientResponse, err error) {
	data, err := l.svcCtx.ClientModel.List(l.ctx)
	if err != nil {
		l.Errorf("Failed to get subscribe application list: %v", err)
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Failed to get subscribe application list")
	}
	var list []types.SubscribeClient
	for _, item := range data {
		var temp types.DownloadLink
		if item.DownloadLink != "" {
			_ = json.Unmarshal([]byte(item.DownloadLink), &temp)
		}
		list = append(list, types.SubscribeClient{
			Id:           item.Id,
			Name:         item.Name,
			Description:  item.Description,
			Icon:         item.Icon,
			Scheme:       item.Scheme,
			IsDefault:    item.IsDefault,
			DownloadLink: temp,
		})
	}
	resp = &types.GetSubscribeClientResponse{
		Total: int64(len(list)),
		List:  list,
	}
	return
}
