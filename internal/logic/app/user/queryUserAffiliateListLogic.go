package user

import (
	"context"

	"github.com/perfect-panel/server/pkg/constant"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type QueryUserAffiliateListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query User Affiliate List
func NewQueryUserAffiliateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserAffiliateListLogic {
	return &QueryUserAffiliateListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryUserAffiliateListLogic) QueryUserAffiliateList(req *types.QueryUserAffiliateListRequest) (resp *types.QueryUserAffiliateListResponse, err error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	var data []*user.User
	var total int64
	err = l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		return db.Model(&user.User{}).Order("id desc").Where("referer_id = ?", u.Id).Count(&total).Limit(req.Size).Offset((req.Page - 1) * req.Size).Find(&data).Error
	})
	if err != nil {
		l.Errorw("Query User Affiliate List failed: %v", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Query User Affiliate List failed: %v", err.Error())
	}

	list := make([]types.UserAffiliate, 0)
	for _, item := range data {
		list = append(list, types.UserAffiliate{
			//Email:        tool.MaskEmail(item.Email),
			Avatar:       item.Avatar,
			RegisteredAt: item.CreatedAt.UnixMilli(),
			Enable:       *item.Enable,
		})
	}
	return &types.QueryUserAffiliateListResponse{
		Total: total,
		List:  list,
	}, nil
}
