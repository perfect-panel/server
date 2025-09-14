package server

import (
	"context"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ResetSortWithNodeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewResetSortWithNodeLogic Reset node sort
func NewResetSortWithNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetSortWithNodeLogic {
	return &ResetSortWithNodeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetSortWithNodeLogic) ResetSortWithNode(req *types.ResetSortRequest) error {
	err := l.svcCtx.NodeModel.Transaction(l.ctx, func(db *gorm.DB) error {
		// find all servers id
		var existingIDs []int64
		db.Model(&node.Node{}).Select("id").Find(&existingIDs)
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
		var servers []*node.Node
		db.Model(&node.Node{}).Order("sort ASC").Find(&servers)
		// create a map of the current sort
		currentSortMap := make(map[int64]int64)
		for _, item := range servers {
			currentSortMap[item.Id] = int64(item.Sort)
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
			s, err := l.svcCtx.NodeModel.FindOneNode(l.ctx, item.Id)
			if err != nil {
				return err
			}
			s.Sort = int(item.Sort)
			if err = l.svcCtx.NodeModel.UpdateNode(l.ctx, s, db); err != nil {
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
