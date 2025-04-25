package server

import (
	"context"
	"strings"

	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type GetNodeTagListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get node tag list
func NewGetNodeTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodeTagListLogic {
	return &GetNodeTagListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNodeTagListLogic) GetNodeTagList() (resp *types.GetNodeTagListResponse, err error) {
	var nodeTags, tags []string
	err = l.svcCtx.ServerModel.Transaction(l.ctx, func(db *gorm.DB) error {

		return db.Model(&server.Server{}).Select("tags").Pluck("tags", &nodeTags).Error
	})

	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get node tag list failed, %s", err.Error())
	}

	for _, tag := range nodeTags {
		tags = append(tags, strings.Split(tag, ",")...)
	}

	// Remove duplicate tags
	tags = tool.RemoveDuplicateElements(tags...)

	return &types.GetNodeTagListResponse{
		Tags: tags,
	}, nil
}
