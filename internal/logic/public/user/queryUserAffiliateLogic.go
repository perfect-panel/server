package user

import (
	"context"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/pkg/constant"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type QueryUserAffiliateLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query User Balance Log
func NewQueryUserAffiliateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserAffiliateLogic {
	return &QueryUserAffiliateLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryUserAffiliateLogic) QueryUserAffiliate() (resp *types.QueryUserAffiliateCountResponse, err error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	var sum int64
	var total int64
	err = l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		return db.Model(&user.User{}).Where("referer_id = ?", u.Id).Count(&total).Find(&user.User{}).Error
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Query User Affiliate failed: %v", err)
	}
	data, _, err := l.svcCtx.LogModel.FilterSystemLog(l.ctx, &log.FilterParams{
		Page:     1,
		Size:     99999,
		Type:     log.TypeCommission.Uint8(),
		ObjectID: u.Id,
	})

	for _, datum := range data {
		content := log.Commission{}
		if err = content.Unmarshal([]byte(datum.Content)); err != nil {
			l.Errorf("[QueryUserAffiliate] unmarshal comission log failed: %v", err.Error())
			continue
		}
		sum += content.Amount
	}

	return &types.QueryUserAffiliateCountResponse{
		Registers:       total,
		TotalCommission: sum,
	}, nil
}
