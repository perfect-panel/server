package user

import (
	"context"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type CreateUserAuthMethodLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create user auth method
func NewCreateUserAuthMethodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserAuthMethodLogic {
	return &CreateUserAuthMethodLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserAuthMethodLogic) CreateUserAuthMethod(req *types.CreateUserAuthMethodRequest) error {
	err := l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		var data *user.AuthMethods
		if err := db.Model(&user.AuthMethods{}).Where("`user_id` = ? AND `auth_type` = ?", req.UserId, req.AuthType).First(&data).Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		data.UserId = req.UserId
		data.AuthType = req.AuthType
		data.AuthIdentifier = req.AuthIdentifier
		if err := db.Model(&user.AuthMethods{}).Save(&data).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorw("[CreateUserAuthMethodLogic] Create User Auth Method Error:", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "Create User Auth Method Error")
	}
	return nil
}
