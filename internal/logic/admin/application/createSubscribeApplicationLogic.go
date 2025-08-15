package application

import (
	"context"

	"github.com/perfect-panel/server/internal/model/client"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type CreateSubscribeApplicationLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateSubscribeApplicationLogic Create subscribe application
func NewCreateSubscribeApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSubscribeApplicationLogic {
	return &CreateSubscribeApplicationLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSubscribeApplicationLogic) CreateSubscribeApplication(req *types.CreateSubscribeApplicationRequest) (resp *types.SubscribeApplication, err error) {
	var link client.DownloadLink
	tool.DeepCopy(&link, req.DownloadLink)
	linkData, err := link.Marshal()
	if err != nil {
		l.Errorf("Failed to marshal download link: %v", err)
		return nil, errors.Wrap(xerr.NewErrCode(xerr.ERROR), " Failed to marshal download link")
	}
	data := &client.SubscribeApplication{
		Name:              req.Name,
		Icon:              req.Icon,
		Description:       req.Description,
		Scheme:            req.Scheme,
		UserAgent:         req.UserAgent,
		IsDefault:         req.IsDefault,
		SubscribeTemplate: req.SubscribeTemplate,
		OutputFormat:      req.OutputFormat,
		DownloadLink:      string(linkData),
	}

	err = l.svcCtx.ClientModel.Insert(l.ctx, data)
	if err != nil {
		l.Errorf("Failed to create subscribe application: %v", err)
		return nil, errors.Wrap(xerr.NewErrCode(xerr.DatabaseInsertError), "Failed to create subscribe application")
	}

	resp = &types.SubscribeApplication{}
	tool.DeepCopy(resp, data)
	resp.DownloadLink = req.DownloadLink

	return
}
