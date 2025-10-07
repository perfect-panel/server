package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/model/node"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
)

type GetServerConfigLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

// NewGetServerConfigLogic Get server config
func NewGetServerConfigLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *GetServerConfigLogic {
	return &GetServerConfigLogic{
		Logger: logger.WithContext(ctx.Request.Context()),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerConfigLogic) GetServerConfig(req *types.GetServerConfigRequest) (resp *types.GetServerConfigResponse, err error) {
	cacheKey := fmt.Sprintf("%s%d:%s", node.ServerConfigCacheKey, req.ServerId, req.Protocol)
	cache, err := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
	if err == nil {
		if cache != "" {
			etag := tool.GenerateETag([]byte(cache))
			//  Check If-None-Match header
			match := l.ctx.GetHeader("If-None-Match")
			if match == etag {
				return nil, xerr.StatusNotModified
			}
			l.ctx.Header("ETag", etag)
			resp = &types.GetServerConfigResponse{}
			err = json.Unmarshal([]byte(cache), resp)
			if err != nil {
				l.Errorw("[ServerConfigCacheKey] json unmarshal error", logger.Field("error", err.Error()))
				return nil, err
			}
			return resp, nil
		}
	}
	data, err := l.svcCtx.NodeModel.FindOneServer(l.ctx, req.ServerId)
	if err != nil {
		l.Errorw("[GetServerConfig] FindOne error", logger.Field("error", err.Error()))
		return nil, err
	}

	// compatible hysteria2, remove in future versions
	protocolRequest := req.Protocol
	if protocolRequest == Hysteria2 {
		protocolRequest = Hysteria
	}

	protocols, err := data.UnmarshalProtocols()
	if err != nil {
		return nil, err
	}
	var cfg map[string]interface{}
	for _, protocol := range protocols {
		if protocol.Type == protocolRequest {
			cfg = l.compatible(protocol)
			break
		}
	}

	resp = &types.GetServerConfigResponse{
		Basic: types.ServerBasic{
			PullInterval: l.svcCtx.Config.Node.NodePullInterval,
			PushInterval: l.svcCtx.Config.Node.NodePushInterval,
		},
		Protocol: req.Protocol,
		Config:   cfg,
	}
	c, err := json.Marshal(resp)
	if err != nil {
		l.Errorw("[GetServerConfig] json marshal error", logger.Field("error", err.Error()))
		return nil, err
	}
	etag := tool.GenerateETag(c)
	l.ctx.Header("ETag", etag)
	if err = l.svcCtx.Redis.Set(l.ctx, cacheKey, c, -1).Err(); err != nil {
		l.Errorw("[GetServerConfig] redis set error", logger.Field("error", err.Error()))
	}
	//  Check If-None-Match header
	match := l.ctx.GetHeader("If-None-Match")
	if match == etag {
		return nil, xerr.StatusNotModified
	}

	return resp, nil
}

func (l *GetServerConfigLogic) compatible(config node.Protocol) map[string]interface{} {
	var result interface{}
	switch config.Type {
	case ShadowSocks:
		result = ShadowsocksNode{
			Port:      config.Port,
			Cipher:    config.Cipher,
			ServerKey: base64.StdEncoding.EncodeToString([]byte(config.ServerKey)),
		}
	case Vless:
		result = VlessNode{
			Port:    config.Port,
			Flow:    config.Flow,
			Network: config.Transport,
			TransportConfig: &TransportConfig{
				Path:                 config.Path,
				Host:                 config.Host,
				ServiceName:          config.ServiceName,
				DisableSNI:           config.DisableSNI,
				ReduceRtt:            config.ReduceRtt,
				UDPRelayMode:         config.UDPRelayMode,
				CongestionController: config.CongestionController,
			},
			Security: config.Security,
			SecurityConfig: &SecurityConfig{
				SNI:                  config.SNI,
				AllowInsecure:        &config.AllowInsecure,
				Fingerprint:          config.Fingerprint,
				RealityServerAddress: config.RealityServerAddr,
				RealityServerPort:    config.RealityServerPort,
				RealityPrivateKey:    config.RealityPrivateKey,
				RealityPublicKey:     config.RealityPublicKey,
				RealityShortId:       config.RealityShortId,
			},
		}
	case Vmess:
		result = VmessNode{
			Port:    config.Port,
			Network: config.Transport,
			TransportConfig: &TransportConfig{
				Path:                 config.Path,
				Host:                 config.Host,
				ServiceName:          config.ServiceName,
				DisableSNI:           config.DisableSNI,
				ReduceRtt:            config.ReduceRtt,
				UDPRelayMode:         config.UDPRelayMode,
				CongestionController: config.CongestionController,
			},
			Security: config.Security,
			SecurityConfig: &SecurityConfig{
				SNI:                  config.SNI,
				AllowInsecure:        &config.AllowInsecure,
				Fingerprint:          config.Fingerprint,
				RealityServerAddress: config.RealityServerAddr,
				RealityServerPort:    config.RealityServerPort,
				RealityPrivateKey:    config.RealityPrivateKey,
				RealityPublicKey:     config.RealityPublicKey,
				RealityShortId:       config.RealityShortId,
			},
		}
	case Trojan:
		result = TrojanNode{
			Port:    config.Port,
			Network: config.Transport,
			TransportConfig: &TransportConfig{
				Path:                 config.Path,
				Host:                 config.Host,
				ServiceName:          config.ServiceName,
				DisableSNI:           config.DisableSNI,
				ReduceRtt:            config.ReduceRtt,
				UDPRelayMode:         config.UDPRelayMode,
				CongestionController: config.CongestionController,
			},
			Security: config.Security,
			SecurityConfig: &SecurityConfig{
				SNI:                  config.SNI,
				AllowInsecure:        &config.AllowInsecure,
				Fingerprint:          config.Fingerprint,
				RealityServerAddress: config.RealityServerAddr,
				RealityServerPort:    config.RealityServerPort,
				RealityPrivateKey:    config.RealityPrivateKey,
				RealityPublicKey:     config.RealityPublicKey,
				RealityShortId:       config.RealityShortId,
			},
		}
	case AnyTLS:
		result = AnyTLSNode{
			Port: config.Port,
			SecurityConfig: &SecurityConfig{
				SNI:                  config.SNI,
				AllowInsecure:        &config.AllowInsecure,
				Fingerprint:          config.Fingerprint,
				RealityServerAddress: config.RealityServerAddr,
				RealityServerPort:    config.RealityServerPort,
				RealityPrivateKey:    config.RealityPrivateKey,
				RealityPublicKey:     config.RealityPublicKey,
				RealityShortId:       config.RealityShortId,
			},
		}
	case Tuic:
		result = TuicNode{
			Port: config.Port,
			SecurityConfig: &SecurityConfig{
				SNI:                  config.SNI,
				AllowInsecure:        &config.AllowInsecure,
				Fingerprint:          config.Fingerprint,
				RealityServerAddress: config.RealityServerAddr,
				RealityServerPort:    config.RealityServerPort,
				RealityPrivateKey:    config.RealityPrivateKey,
				RealityPublicKey:     config.RealityPublicKey,
				RealityShortId:       config.RealityShortId,
			},
		}
	case Hysteria:
		result = Hysteria2Node{
			Port:         config.Port,
			HopPorts:     config.HopPorts,
			HopInterval:  config.HopInterval,
			ObfsPassword: config.ObfsPassword,
			SecurityConfig: &SecurityConfig{
				SNI:                  config.SNI,
				AllowInsecure:        &config.AllowInsecure,
				Fingerprint:          config.Fingerprint,
				RealityServerAddress: config.RealityServerAddr,
				RealityServerPort:    config.RealityServerPort,
				RealityPrivateKey:    config.RealityPrivateKey,
				RealityPublicKey:     config.RealityPublicKey,
				RealityShortId:       config.RealityShortId,
			},
		}

	}
	var resp map[string]interface{}
	s, _ := json.Marshal(result)
	_ = json.Unmarshal(s, &resp)
	return resp
}
