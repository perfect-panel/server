package system

import (
	"context"
	"strings"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
)

type GetModuleConfigLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get Module Config
func NewGetModuleConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetModuleConfigLogic {
	return &GetModuleConfigLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetModuleConfigLogic) GetModuleConfig() (resp *types.ModuleConfig, err error) {
	//value, exists := os.LookupEnv("SECRET_KEY")
	//if !exists {
	//	return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), " SECRET_KEY not set in environment variables")
	//}

	return &types.ModuleConfig{
		//Secret:         value,
		ServiceName:    constant.ServiceName,
		ServiceVersion: strings.ReplaceAll(constant.Version, "v", ""),
	}, nil
}
