package subscribe

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/perfect-panel/server/adapter"
	"github.com/perfect-panel/server/internal/model/client"
	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/node"

	"github.com/perfect-panel/server/internal/model/user"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

//goland:noinspection GoNameStartsWithPackageName
type SubscribeLogic struct {
	ctx *gin.Context
	svc *svc.ServiceContext
	logger.Logger
}

func NewSubscribeLogic(ctx *gin.Context, svc *svc.ServiceContext) *SubscribeLogic {
	return &SubscribeLogic{
		ctx:    ctx,
		svc:    svc,
		Logger: logger.WithContext(ctx.Request.Context()),
	}
}

func (l *SubscribeLogic) Handler(req *types.SubscribeRequest) (resp *types.SubscribeResponse, err error) {
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
			// Special handling for Stash
			if strings.Contains(userAgent, "stash") && !strings.Contains(u, "stash") {
				continue
			}
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
	subscribeStatus = true
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

func (l *SubscribeLogic) getUserSubscribe(token string) (*user.Subscribe, error) {
	userSub, err := l.svc.UserModel.FindOneSubscribeByToken(l.ctx.Request.Context(), token)
	if err != nil {
		l.Infow("[Generate Subscribe]find subscribe error: %v", logger.Field("error", err.Error()), logger.Field("token", token))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe error: %v", err.Error())
	}

	//  Ignore expiration check
	//if userSub.Status > 1 {
	//	l.Infow("[Generate Subscribe]subscribe is not available", logger.Field("status", int(userSub.Status)), logger.Field("token", token))
	//	return nil, errors.Wrapf(xerr.NewErrCode(xerr.SubscribeNotAvailable), "subscribe is not available")
	//}

	return userSub, nil
}

func (l *SubscribeLogic) logSubscribeActivity(subscribeStatus bool, userSub *user.Subscribe, req *types.SubscribeRequest) {
	if !subscribeStatus {
		return
	}

	subscribeLog := log.Subscribe{
		Token:           req.Token,
		UserAgent:       req.UA,
		ClientIP:        l.ctx.ClientIP(),
		UserSubscribeId: userSub.Id,
	}

	content, _ := subscribeLog.Marshal()

	err := l.svc.LogModel.Insert(l.ctx.Request.Context(), &log.SystemLog{
		Type:     log.TypeSubscribe.Uint8(),
		ObjectID: userSub.UserId, // log user id
		Date:     time.Now().Format(time.DateOnly),
		Content:  string(content),
	})
	if err != nil {
		l.Errorw("[Generate Subscribe]insert subscribe log error: %v", logger.Field("error", err.Error()))
	}
}

func (l *SubscribeLogic) getServers(userSub *user.Subscribe) ([]*node.Node, error) {
	if l.isSubscriptionExpired(userSub) {
		return l.createExpiredServers(), nil
	}

	subDetails, err := l.svc.SubscribeModel.FindOne(l.ctx.Request.Context(), userSub.SubscribeId)
	if err != nil {
		l.Errorw("[Generate Subscribe]find subscribe details error: %v", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe details error: %v", err.Error())
	}

	nodeIds := tool.StringToInt64Slice(subDetails.Nodes)
	tags := strings.Split(subDetails.NodeTags, ",")

	l.Debugf("[Generate Subscribe]nodes: %v, NodeTags: %v", nodeIds, tags)

	enable := true

	_, nodes, err := l.svc.NodeModel.FilterNodeList(l.ctx.Request.Context(), &node.FilterNodeParams{
		Page:    1,
		Size:    1000,
		NodeId:  nodeIds,
		Tag:     tool.RemoveDuplicateElements(tags...),
		Preload: true,
		Enabled: &enable, // Only get enabled nodes
	})

	l.Debugf("[Query Subscribe]found servers: %v", len(nodes))

	if err != nil {
		l.Errorw("[Generate Subscribe]find server details error: %v", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find server details error: %v", err.Error())
	}
	logger.Debugf("[Generate Subscribe]found servers: %v", len(nodes))
	return nodes, nil
}

func (l *SubscribeLogic) isSubscriptionExpired(userSub *user.Subscribe) bool {
	return userSub.ExpireTime.Unix() < time.Now().Unix() && userSub.ExpireTime.Unix() != 0
}

func (l *SubscribeLogic) createExpiredServers() []*node.Node {
	enable := true
	host := l.getFirstHostLine()

	return []*node.Node{
		{
			Name:    "Subscribe Expired",
			Tags:    "",
			Port:    18080,
			Address: "127.0.0.1",
			Server: &node.Server{
				Id:        1,
				Name:      "Subscribe Expired",
				Protocols: "[{\"type\":\"shadowsocks\",\"cipher\":\"aes-256-gcm\",\"port\":1}]",
			},
			Protocol: "shadowsocks",
			Enabled:  &enable,
		},
		{
			Name:    host,
			Tags:    "",
			Port:    18080,
			Address: "127.0.0.1",
			Server: &node.Server{
				Id:        1,
				Name:      "Subscribe Expired",
				Protocols: "[{\"type\":\"shadowsocks\",\"cipher\":\"aes-256-gcm\",\"port\":1}]",
			},
			Protocol: "shadowsocks",
			Enabled:  &enable,
		},
	}
}

func (l *SubscribeLogic) getFirstHostLine() string {
	host := l.svc.Config.Host
	lines := strings.Split(host, "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return host
}
