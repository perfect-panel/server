package application

import (
	"context"
	"time"

	"github.com/perfect-panel/server/adapter"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type PreviewSubscribeTemplateLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Preview Template
func NewPreviewSubscribeTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewSubscribeTemplateLogic {
	return &PreviewSubscribeTemplateLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreviewSubscribeTemplateLogic) PreviewSubscribeTemplate(req *types.PreviewSubscribeTemplateRequest) (resp *types.PreviewSubscribeTemplateResponse, err error) {
	servers, err := l.svcCtx.ServerModel.FindAllServer(l.ctx)
	if err != nil {
		l.Errorf("[PreviewSubscribeTemplateLogic] FindAllServer error: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FindAllServer error: %v", err.Error())
	}

	data, err := l.svcCtx.ClientModel.FindOne(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[PreviewSubscribeTemplateLogic] FindOne error: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FindOneClient error: %v", err.Error())
	}

	sub := adapter.NewAdapter(data.SubscribeTemplate, adapter.WithServers(servers),
		adapter.WithSiteName("PerfectPanel"),
		adapter.WithSubscribeName("Test Subscribe"),
		adapter.WithOutputFormat(data.OutputFormat),
		adapter.WithUserInfo(adapter.User{
			Password:     "test-password",
			ExpiredAt:    time.Now().AddDate(1, 0, 0),
			Download:     0,
			Upload:       0,
			Traffic:      1000,
			SubscribeURL: "https://example.com/subscribe",
		}))
	// Get client config
	a, err := sub.Client()
	if err != nil {
		l.Errorf("[PreviewSubscribeTemplateLogic] Client error: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrMsg(err.Error()), "Client error: %v", err.Error())
	}
	bytes, err := a.Build()
	if err != nil {
		l.Errorf("[PreviewSubscribeTemplateLogic] Build error: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrMsg(err.Error()), "Build error: %v", err.Error())
	}
	return &types.PreviewSubscribeTemplateResponse{
		Template: string(bytes),
	}, nil
}
