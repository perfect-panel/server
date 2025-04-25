package user

import (
	"context"

	"github.com/perfect-panel/ppanel-server/pkg/constant"

	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
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
	err = l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		return db.Model(&user.CommissionLog{}).
			Where("user_id = ?", u.Id).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&sum).Error
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Query User Affiliate failed: %v", err)
	}

	return &types.QueryUserAffiliateCountResponse{
		Registers:       total,
		TotalCommission: sum,
	}, nil
}
