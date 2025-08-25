package log

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type FilterUserSubscribeTrafficLogLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFilterUserSubscribeTrafficLogLogic Filter user subscribe traffic log
func NewFilterUserSubscribeTrafficLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilterUserSubscribeTrafficLogLogic {
	return &FilterUserSubscribeTrafficLogLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FilterUserSubscribeTrafficLogLogic) FilterUserSubscribeTrafficLog(req *types.FilterSubscribeTrafficRequest) (resp *types.FilterSubscribeTrafficResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
