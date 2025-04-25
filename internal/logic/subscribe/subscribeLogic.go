package subscribe

import (
	"fmt"
	"strings"
	"time"

	"github.com/perfect-panel/server/pkg/adapter"
	"github.com/perfect-panel/server/pkg/adapter/shadowrocket"
	"github.com/perfect-panel/server/pkg/adapter/surfboard"

	"github.com/perfect-panel/server/internal/model/server"

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

func (l *SubscribeLogic) Generate(req *types.SubscribeRequest) (*types.SubscribeResponse, error) {
	userSub, err := l.getUserSubscribe(req.Token)
	if err != nil {
		return nil, err
	}

	var subscribeStatus = false
	defer func() {
		l.logSubscribeActivity(subscribeStatus, userSub, req)
	}()

	servers, err := l.getServers(userSub)
	if err != nil {
		return nil, err
	}

	rules, err := l.getRules()
	if err != nil {
		return nil, err
	}

	resp, headerInfo, err := l.buildClientConfig(req, userSub, servers, rules)
	if err != nil {
		return nil, err
	}

	subscribeStatus = true
	return &types.SubscribeResponse{
		Config: resp,
		Header: headerInfo,
	}, nil
}

func (l *SubscribeLogic) getUserSubscribe(token string) (*user.Subscribe, error) {
	userSub, err := l.svc.UserModel.FindOneSubscribeByToken(l.ctx.Request.Context(), token)
	if err != nil {
		l.Infow("[Generate Subscribe]find subscribe error: %v", logger.Field("error", err.Error()), logger.Field("token", token))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe error: %v", err.Error())
	}

	if userSub.Status != 1 {
		l.Infow("[Generate Subscribe]subscribe is not available", logger.Field("status", int(userSub.Status)), logger.Field("token", token))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.SubscribeNotAvailable), "subscribe is not available")
	}

	return userSub, nil
}

func (l *SubscribeLogic) logSubscribeActivity(subscribeStatus bool, userSub *user.Subscribe, req *types.SubscribeRequest) {
	if !subscribeStatus {
		return
	}

	err := l.svc.UserModel.InsertSubscribeLog(l.ctx.Request.Context(), &user.SubscribeLog{
		UserId:          userSub.UserId,
		UserSubscribeId: userSub.Id,
		Token:           req.Token,
		IP:              l.ctx.ClientIP(),
		UserAgent:       l.ctx.Request.UserAgent(),
	})
	if err != nil {
		l.Errorw("[Generate Subscribe]insert subscribe log error: %v", logger.Field("error", err.Error()))
	}
}

func (l *SubscribeLogic) getServers(userSub *user.Subscribe) ([]*server.Server, error) {
	if l.isSubscriptionExpired(userSub) {
		return l.createExpiredServers(), nil
	}

	subDetails, err := l.svc.SubscribeModel.FindOne(l.ctx.Request.Context(), userSub.SubscribeId)
	if err != nil {
		l.Errorw("[Generate Subscribe]find subscribe details error: %v", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe details error: %v", err.Error())
	}

	serverIds := tool.StringToInt64Slice(subDetails.Server)
	groupIds := tool.StringToInt64Slice(subDetails.ServerGroup)

	servers, err := l.svc.ServerModel.FindServerDetailByGroupIdsAndIds(l.ctx.Request.Context(), groupIds, serverIds)
	if err != nil {
		l.Errorw("[Generate Subscribe]find server details error: %v", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find server details error: %v", err.Error())
	}

	return servers, nil
}

func (l *SubscribeLogic) isSubscriptionExpired(userSub *user.Subscribe) bool {
	return userSub.ExpireTime.Unix() < time.Now().Unix() && userSub.ExpireTime.Unix() != 0
}

func (l *SubscribeLogic) createExpiredServers() []*server.Server {
	enable := true
	host := l.getFirstHostLine()

	return []*server.Server{
		{
			Name:       "Subscribe Expired",
			ServerAddr: "127.0.0.1",
			RelayMode:  "none",
			Protocol:   "shadowsocks",
			Config:     "{\"method\":\"aes-256-gcm\",\"port\":1}",
			Enable:     &enable,
			Sort:       0,
		},
		{
			Name:       host,
			ServerAddr: "127.0.0.1",
			RelayMode:  "none",
			Protocol:   "shadowsocks",
			Config:     "{\"method\":\"aes-256-gcm\",\"port\":1}",
			Enable:     &enable,
			Sort:       0,
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

func (l *SubscribeLogic) getRules() ([]*server.RuleGroup, error) {
	rules, err := l.svc.ServerModel.QueryAllRuleGroup(l.ctx)
	if err != nil {
		l.Errorw("[Generate Subscribe]find rule group error: %v", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find rule group error: %v", err.Error())
	}
	return rules, nil
}

func (l *SubscribeLogic) buildClientConfig(req *types.SubscribeRequest, userSub *user.Subscribe, servers []*server.Server, rules []*server.RuleGroup) ([]byte, string, error) {
	proxyManager := adapter.NewAdapter(servers, rules)
	clientType := l.getClientType(req)
	var resp []byte
	var err error

	l.Logger.Info(fmt.Sprintf("[Generate Subscribe] %s", clientType), logger.Field("ua", req.UA), logger.Field("flag", req.Flag))

	switch clientType {
	case "clash":
		resp, err = proxyManager.BuildClash(userSub.UUID)
		if err != nil {
			l.Errorw("[Generate Subscribe] build clash error", logger.Field("error", err.Error()))
			return nil, "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "build clash error: %v", err.Error())
		}
		l.setClashHeaders()
	case "sing-box":
		resp, err = proxyManager.BuildSingbox(userSub.UUID)
		if err != nil {
			l.Errorw("[Generate Subscribe] build sing-box error", logger.Field("error", err.Error()))
			return nil, "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "build sing-box error: %v", err.Error())
		}
	case "quantumult":
		resp = []byte(proxyManager.BuildQuantumultX(userSub.UUID))
	case "shadowrocket":
		resp = proxyManager.BuildShadowrocket(userSub.UUID, shadowrocket.UserInfo{
			Upload:       userSub.Upload,
			Download:     userSub.Download,
			TotalTraffic: userSub.Traffic,
			ExpiredDate:  userSub.ExpireTime,
		})
	case "loon":
		resp = proxyManager.BuildLoon(userSub.UUID)
	case "surfboard":
		subsURL := l.getSubscribeURL(userSub.Token)
		resp = proxyManager.BuildSurfboard(l.svc.Config.Site.SiteName, surfboard.UserInfo{
			Upload:       userSub.Upload,
			Download:     userSub.Download,
			TotalTraffic: userSub.Traffic,
			ExpiredDate:  userSub.ExpireTime,
			UUID:         userSub.UUID,
			SubscribeURL: subsURL,
		})
		l.setSurfboardHeaders()
	default:
		resp = proxyManager.BuildGeneral(userSub.UUID)
	}

	headerInfo := fmt.Sprintf("upload=%d;download=%d;total=%d;expire=%d",
		userSub.Upload, userSub.Download, userSub.Traffic, userSub.ExpireTime.Unix())

	return resp, headerInfo, nil
}

func (l *SubscribeLogic) setClashHeaders() {
	l.ctx.Header("content-disposition", fmt.Sprintf("tattachment;filename*=UTF-8''%s.yaml", l.svc.Config.Site.SiteName))
	l.ctx.Header("Profile-Update-Interval", "24")
	l.ctx.Header("Content-Type", "application/octet-stream; charset=UTF-8")
}

func (l *SubscribeLogic) setSurfboardHeaders() {
	l.ctx.Header("content-disposition", fmt.Sprintf("attachment;filename*=UTF-8''%s.conf", l.svc.Config.Site.SiteName))
	l.ctx.Header("Content-Type", "application/octet-stream; charset=UTF-8")
}

func (l *SubscribeLogic) getSubscribeURL(token string) string {
	if l.svc.Config.Subscribe.PanDomain {
		return fmt.Sprintf("https://%s", l.ctx.Request.Host)
	}

	if l.svc.Config.Subscribe.SubscribeDomain != "" {
		domains := strings.Split(l.svc.Config.Subscribe.SubscribeDomain, "\n")
		return fmt.Sprintf("https://%s%s?token=%s&flag=surfboard", domains[0], l.svc.Config.Subscribe.SubscribePath, token)
	}

	return fmt.Sprintf("https://%s%s?token=%s&flag=surfboard", l.ctx.Request.Host, l.svc.Config.Subscribe.SubscribePath, token)
}

func (l *SubscribeLogic) getClientType(req *types.SubscribeRequest) string {
	clientTypeMap := map[string]string{
		"clash":        "clash",
		"meta":         "clash",
		"sing-box":     "sing-box",
		"hiddify":      "sing-box",
		"surge":        "surge",
		"quantumult":   "quantumult",
		"shadowrocket": "shadowrocket",
		"loon":         "loon",
		"surfboard":    "surfboard",
	}

	findClient := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			return ""
		}

		for key, clientType := range clientTypeMap {
			if strings.Contains(s, key) {
				return clientType
			}
		}

		return ""
	}

	// 优先检查Flag参数
	if typ := findClient(req.Flag); typ != "" {
		return typ
	}

	// 其次检查UA参数
	return findClient(req.UA)
}
