package subscribe

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/perfect-panel/server/adapter"
	"github.com/perfect-panel/server/internal/model/client"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

func (l *SubscribeLogic) V2(req *types.SubscribeRequest) (resp *types.SubscribeResponse, err error) {
	// query client list
	clients, err := l.svc.ClientModel.List(l.ctx.Request.Context())
	if err != nil {
		l.Errorw("[SubscribeLogic] Query client list failed", logger.Field("error", err.Error()))
		return nil, err
	}

	userAgent := strings.ToLower(l.ctx.Request.UserAgent())

	var targetApp, defaultApp *client.SubscribeApplication

	for _, item := range clients {
		u := strings.ToLower(item.UserAgent)
		if item.IsDefault {
			defaultApp = item
		}
		if strings.Contains(userAgent, u) {
			targetApp = item
			break
		}
	}
	if targetApp == nil {
		l.Debugf("[SubscribeLogic] No matching client found", logger.Field("userAgent", userAgent))
		if defaultApp == nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "No matching client found for user agent: %s", userAgent)
		}
		targetApp = defaultApp
	}
	// Find user subscribe by token
	userSubscribe, err := l.getUserSubscribe(req.Token)
	if err != nil {
		l.Errorw("[SubscribeLogic] Get user subscribe failed", logger.Field("error", err.Error()), logger.Field("token", req.Token))
		return nil, err
	}

	var subscribeStatus = false
	defer func() {
		l.logSubscribeActivity(subscribeStatus, userSubscribe, req)
	}()
	// find subscribe info
	subscribeInfo, err := l.svc.SubscribeModel.FindOne(l.ctx.Request.Context(), userSubscribe.SubscribeId)
	if err != nil {
		l.Errorw("[SubscribeLogic] Find subscribe info failed", logger.Field("error", err.Error()), logger.Field("subscribeId", userSubscribe.SubscribeId))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Find subscribe info failed: %v", err.Error())
	}

	// Find server list by user subscribe
	servers, err := l.getServers(userSubscribe)
	if err != nil {
		return nil, err
	}
	a := adapter.NewAdapter(
		targetApp.SubscribeTemplate,
		adapter.WithServers(servers),
		adapter.WithSiteName(l.svc.Config.Site.SiteName),
		adapter.WithSubscribeName(subscribeInfo.Name),
		adapter.WithOutputFormat(targetApp.OutputFormat),
		adapter.WithUserInfo(adapter.User{
			Password:     userSubscribe.UUID,
			ExpiredAt:    userSubscribe.ExpireTime,
			Download:     userSubscribe.Download,
			Upload:       userSubscribe.Upload,
			Traffic:      userSubscribe.Traffic,
			SubscribeURL: l.getSubscribeV2URL(req.Token),
		}),
	)

	// Get client config
	adapterClient, err := a.Client()
	if err != nil {
		l.Errorw("[SubscribeLogic] Client error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(500), "Client error: %v", err.Error())
	}
	bytes, err := adapterClient.Build()
	if err != nil {
		l.Errorw("[SubscribeLogic] Build client config failed", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(500), "Build client config failed: %v", err.Error())
	}

	var formats = []string{"json", "yaml", "conf"}

	for _, format := range formats {
		if format == strings.ToLower(targetApp.OutputFormat) {
			l.ctx.Header("content-disposition", fmt.Sprintf("attachment;filename*=UTF-8''%s.%s", url.QueryEscape(l.svc.Config.Site.SiteName), format))
			l.ctx.Header("Content-Type", "application/octet-stream; charset=UTF-8")

		}
	}

	resp = &types.SubscribeResponse{
		Config: bytes,
		Header: fmt.Sprintf(
			"upload=%d;download=%d;total=%d;expire=%d",
			userSubscribe.Upload, userSubscribe.Download, userSubscribe.Traffic, userSubscribe.ExpireTime.Unix(),
		),
	}
	return
}

func (l *SubscribeLogic) getSubscribeV2URL(token string) string {
	if l.svc.Config.Subscribe.PanDomain {
		return fmt.Sprintf("https://%s", l.ctx.Request.Host)
	}

	if l.svc.Config.Subscribe.SubscribeDomain != "" {
		domains := strings.Split(l.svc.Config.Subscribe.SubscribeDomain, "\n")
		return fmt.Sprintf("https://%s%s?token=%s", domains[0], l.svc.Config.Subscribe.SubscribePath, token)
	}

	return fmt.Sprintf("https://%s%s?token=%s&", l.ctx.Request.Host, l.svc.Config.Subscribe.SubscribePath, token)
}
