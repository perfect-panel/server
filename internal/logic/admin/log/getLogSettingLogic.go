package log

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
)

type GetLogSettingLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get log setting
func NewGetLogSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogSettingLogic {
	return &GetLogSettingLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLogSettingLogic) GetLogSetting() (resp *types.LogSetting, err error) {
	configs, err := l.svcCtx.SystemModel.GetLogConfig(l.ctx)
	if err != nil {
		l.Errorw("[GetLogSetting] Database query error", logger.Field("error", err.Error()))
		return nil, err
	}
	resp = &types.LogSetting{}
	// reflect to response
	tool.SystemConfigSliceReflectToStruct(configs, resp)
	return
}
