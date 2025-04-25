package server

import (
	"context"

	"github.com/perfect-panel/ppanel-server/internal/model/server"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
)

type NodeSortLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Node sort
func NewNodeSortLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NodeSortLogic {
	return &NodeSortLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NodeSortLogic) NodeSort(req *types.NodeSortRequest) error {
	err := l.svcCtx.ServerModel.Transaction(l.ctx, func(db *gorm.DB) error {
		// find all servers id
		var existingIDs []int64
		db.Model(&server.Server{}).Select("id").Find(&existingIDs)
		// check if the id is valid
		validIDMap := make(map[int64]bool)
		for _, id := range existingIDs {
			validIDMap[id] = true
		}
		// check if the sort is valid
		var validItems []types.SortItem
		for _, item := range req.Sort {
			if validIDMap[item.Id] {
				validItems = append(validItems, item)
			}
		}
		// query all servers
		var servers []*server.Server
		db.Model(&server.Server{}).Order("sort ASC").Find(&servers)
		// create a map of the current sort
		currentSortMap := make(map[int64]int64)
		for _, item := range servers {
			currentSortMap[item.Id] = item.Sort
		}

		// new sort map
		newSortMap := make(map[int64]int64)
		for _, item := range validItems {
			newSortMap[item.Id] = item.Sort
		}

		var itemsToUpdate []types.SortItem
		for _, item := range validItems {
			if oldSort, exists := currentSortMap[item.Id]; exists && oldSort != item.Sort {
				itemsToUpdate = append(itemsToUpdate, item)
			}
		}
		for _, item := range itemsToUpdate {
			if err := db.Model(&server.Server{}).Where("id = ?", item.Id).Update("sort", item.Sort).Error; err != nil {
				l.Errorw("[NodeSort] Update Database Error: ", logger.Field("error", err.Error()), logger.Field("id", item.Id), logger.Field("sort", item.Sort))
				return err
			}
		}
		return nil
	})
	if err != nil {
		l.Errorw("[NodeSort] Update Database Error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), err.Error())
	}
	return nil
}
