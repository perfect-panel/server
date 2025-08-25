package log

import (
	"context"
	"time"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type FilterServerTrafficLogLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFilterServerTrafficLogLogic Filter server traffic log
func NewFilterServerTrafficLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilterServerTrafficLogLogic {
	return &FilterServerTrafficLogLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FilterServerTrafficLogLogic) FilterServerTrafficLog(req *types.FilterServerTrafficLogRequest) (resp *types.FilterServerTrafficLogResponse, err error) {
	today := time.Now().Format("2006-01-02")
	if req.Date == "" || req.Date == today {
		return l.handlerToday(req)
	} else {
		return l.handlerSpecify(req)
	}
}

func (l *FilterServerTrafficLogLogic) handlerToday(req *types.FilterServerTrafficLogRequest) (resp *types.FilterServerTrafficLogResponse, err error) {

	return
}

func (l *FilterServerTrafficLogLogic) handlerSpecify(req *types.FilterServerTrafficLogRequest) (resp *types.FilterServerTrafficLogResponse, err error) {
	return
}
