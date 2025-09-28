package system

import (
	"context"
	"reflect"

	"github.com/perfect-panel/server/initialize"
	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/model/system"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
)

type UpdateNodeConfigLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNodeConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNodeConfigLogic {
	return &UpdateNodeConfigLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNodeConfigLogic) UpdateNodeConfig(req *types.NodeConfig) error {
	v := reflect.ValueOf(*req)
	// Get the reflection type of the structure
	t := v.Type()
	err := l.svcCtx.SystemModel.Transaction(l.ctx, func(db *gorm.DB) error {
		var err error
		for i := 0; i < v.NumField(); i++ {
			// Get the field name
			fieldName := t.Field(i).Name
			// Get the field value to string
			var fieldValue string

			fieldValue = tool.ConvertValueToString(v.Field(i))
			// Update the server config
			err = db.Model(&system.System{}).Where("`category` = 'server' and `key` = ?", fieldName).Update("value", fieldValue).Error
			if err != nil {
				break
			}
		}
		return l.svcCtx.Redis.Del(l.ctx, config.NodeConfigKey).Err()
	})
	if err != nil {
		l.Errorw("[UpdateNodeConfig] update node config error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update server config error: %v", err)
	}
	initialize.Node(l.svcCtx)
	return nil
}
