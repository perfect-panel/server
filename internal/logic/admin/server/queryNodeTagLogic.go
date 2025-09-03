package server

import (
	"context"
	"strings"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryNodeTagLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewQueryNodeTagLogic Query all node tags
func NewQueryNodeTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryNodeTagLogic {
	return &QueryNodeTagLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryNodeTagLogic) QueryNodeTag() (resp *types.QueryNodeTagResponse, err error) {

	var nodes []*node.Node
	if err = l.svcCtx.DB.WithContext(l.ctx).Model(&node.Node{}).Find(&nodes).Error; err != nil {
		l.Errorw("[QueryNodeTag] Query Database Error: ", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "[QueryNodeTag] Query Database Error")
	}
	var tags []string
	for _, item := range nodes {
		tags = append(tags, strings.Split(item.Tags, ",")...)
	}

	return &types.QueryNodeTagResponse{
		Tags: tool.RemoveDuplicateElements(tags...),
	}, nil
}
