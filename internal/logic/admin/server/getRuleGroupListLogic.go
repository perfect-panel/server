package server

import (
	"context"
	"strings"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetRuleGroupListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get rule group list
func NewGetRuleGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRuleGroupListLogic {
	return &GetRuleGroupListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRuleGroupListLogic) GetRuleGroupList() (resp *types.GetRuleGroupResponse, err error) {
	nodeRuleGroupList, err := l.svcCtx.ServerModel.QueryAllRuleGroup(l.ctx)
	if err != nil {
		l.Errorw("[GetRuleGroupList] Query Database Error: ", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), err.Error())
	}
	nodeRuleGroups := make([]types.ServerRuleGroup, len(nodeRuleGroupList))
	for i, v := range nodeRuleGroupList {
		nodeRuleGroups[i] = types.ServerRuleGroup{
			Id:        v.Id,
			Icon:      v.Icon,
			Name:      v.Name,
			Tags:      strings.Split(v.Tags, ","),
			Rules:     v.Rules,
			Enable:    v.Enable,
			CreatedAt: v.CreatedAt.UnixMilli(),
			UpdatedAt: v.UpdatedAt.UnixMilli(),
		}
	}
	return &types.GetRuleGroupResponse{
		Total: int64(len(nodeRuleGroups)),
		List:  nodeRuleGroups,
	}, nil
}
