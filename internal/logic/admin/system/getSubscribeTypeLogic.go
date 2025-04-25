package system

import (
	"context"

	"github.com/perfect-panel/server/internal/model/subscribeType"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetSubscribeTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger.Logger
}

func NewGetSubscribeTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubscribeTypeLogic {
	return &GetSubscribeTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logger.WithContext(ctx),
	}
}

func (l *GetSubscribeTypeLogic) GetSubscribeType() (resp *types.SubscribeType, err error) {
	var list []*subscribeType.SubscribeType
	err = l.svcCtx.DB.Model(&subscribeType.SubscribeType{}).Find(&list).Error
	if err != nil {
		l.Errorw("[GetSubscribeType] get subscribe type failed", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get subscribe type failed: %v", err)
	}
	typeList := make([]string, 0)
	for _, item := range list {
		typeList = append(typeList, item.Name)
	}
	return &types.SubscribeType{
		SubscribeTypes: typeList,
	}, nil
}
