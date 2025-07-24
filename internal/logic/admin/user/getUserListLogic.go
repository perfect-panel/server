package user

import (
	"context"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/phone"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger.Logger
}

func NewGetUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListLogic {
	return &GetUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logger.WithContext(ctx),
	}
}
func (l *GetUserListLogic) GetUserList(req *types.GetUserListRequest) (*types.GetUserListResponse, error) {
	list, total, err := l.svcCtx.UserModel.QueryPageList(l.ctx, req.Page, req.Size, &user.UserFilterParams{
		UserId:          req.UserId,
		Search:          req.Search,
		SubscribeId:     req.SubscribeId,
		UserSubscribeId: req.UserSubscribeId,
		Order:           "DESC",
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "GetUserListLogic failed: %v", err.Error())
	}

	userRespList := make([]types.User, 0, len(list))

	for _, item := range list {
		var user types.User
		tool.DeepCopy(&user, item)

		// 处理 AuthMethods
		authMethods := make([]types.UserAuthMethod, len(user.AuthMethods)) // 直接创建目标 slice
		for i, method := range user.AuthMethods {
			tool.DeepCopy(&authMethods[i], method)
			if method.AuthType == "mobile" {
				authMethods[i].AuthIdentifier = phone.FormatToInternational(method.AuthIdentifier)
			}
		}
		user.AuthMethods = authMethods

		userRespList = append(userRespList, user)
	}

	return &types.GetUserListResponse{
		Total: total,
		List:  userRespList,
	}, nil
}
