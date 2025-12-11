package subscribe

import (
	"context"
	"strconv"
	"time"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/uuidx"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type ResetAllSubscribeTokenLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Reset all subscribe tokens
func NewResetAllSubscribeTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetAllSubscribeTokenLogic {
	return &ResetAllSubscribeTokenLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetAllSubscribeTokenLogic) ResetAllSubscribeToken() (resp *types.ResetAllSubscribeTokenResponse, err error) {
	var list []*user.Subscribe
	tx := l.svcCtx.DB.WithContext(l.ctx).Begin()
	// select all active and Finished subscriptions
	if err = tx.Model(&user.Subscribe{}).Where("`status` IN ?", []int64{1, 2}).Find(&list).Error; err != nil {
		logger.Errorf("[ResetAllSubscribeToken] Failed to fetch subscribe list: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Failed to fetch subscribe list: %v", err.Error())
	}

	for _, sub := range list {
		sub.Token = uuidx.SubscribeToken(strconv.FormatInt(time.Now().UnixMilli(), 10) + strconv.FormatInt(sub.Id, 10))
		sub.UUID = uuidx.NewUUID().String()
		if err = tx.Model(&user.Subscribe{}).Where("id = ?", sub.Id).Save(sub).Error; err != nil {
			tx.Rollback()
			logger.Errorf("[ResetAllSubscribeToken] Failed to update subscribe token for ID %d: %v", sub.Id, err.Error())
			return &types.ResetAllSubscribeTokenResponse{
				Success: false,
			}, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "Failed to update subscribe token for ID %d: %v", sub.Id, err.Error())
		}
	}
	if err = tx.Commit().Error; err != nil {
		logger.Errorf("[ResetAllSubscribeToken] Failed to commit transaction: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "Failed to commit transaction: %v", err.Error())
	}

	return &types.ResetAllSubscribeTokenResponse{
		Success: true,
	}, nil
}
