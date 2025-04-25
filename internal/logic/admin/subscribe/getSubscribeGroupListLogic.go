package subscribe

import (
	"context"

	"github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetSubscribeGroupListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get subscribe group list
func NewGetSubscribeGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubscribeGroupListLogic {
	return &GetSubscribeGroupListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubscribeGroupListLogic) GetSubscribeGroupList() (resp *types.GetSubscribeGroupListResponse, err error) {
	var list []*subscribe.Group
	var total int64
	err = l.svcCtx.DB.Model(&subscribe.Group{}).Count(&total).Find(&list).Error
	if err != nil {
		l.Logger.Error("[GetSubscribeGroupListLogic] get subscribe group list failed: ", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get subscribe group list failed: %v", err.Error())
	}
	groupList := make([]types.SubscribeGroup, 0)
	tool.DeepCopy(&groupList, list)
	return &types.GetSubscribeGroupListResponse{
		Total: total,
		List:  groupList,
	}, nil
}
