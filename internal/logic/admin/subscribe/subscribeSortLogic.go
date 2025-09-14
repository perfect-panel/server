package subscribe

import (
	"context"

	"github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type SubscribeSortLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSubscribeSortLogic Subscribe sort
func NewSubscribeSortLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeSortLogic {
	return &SubscribeSortLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubscribeSortLogic) SubscribeSort(req *types.SubscribeSortRequest) error {
	var sort = make(map[int64]int64, len(req.Sort))
	var ids []int64
	for i, v := range req.Sort {
		sort[v.Id] = int64(i)
		ids = append(ids, v.Id)
	}
	// query min sort by ids
	minSort, err := l.svcCtx.SubscribeModel.QuerySubscribeMinSortByIds(l.ctx, ids)
	if err != nil {
		l.Logger.Error("[SubscribeSortLogic] query subscribe list by ids error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query subscribe list by ids error: %v", err.Error())
	}
	_, subs, err := l.svcCtx.SubscribeModel.FilterList(l.ctx, &subscribe.FilterParams{
		Page: 1,
		Size: 9999,
		Ids:  ids,
	})
	if err != nil {
		l.Logger.Error("[SubscribeSortLogic] query subscribe list by ids error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query subscribe list by ids error: %v", err.Error())
	}
	// reordering
	for _, sub := range subs {
		if newSort, ok := sort[sub.Id]; ok {
			sub.Sort = minSort + newSort
		}
	}
	// update sort
	err = l.svcCtx.SubscribeModel.Transaction(l.ctx, func(db *gorm.DB) error {
		return db.Save(subs).Error
	})
	if err != nil {
		l.Logger.Error("[SubscribeSortLogic] update subscribe sort error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update subscribe sort error: %v", err.Error())
	}
	l.Logger.Info("[UpdateSubscribeSort] Successfully updated subscribe sort")
	return nil
}
