package server

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/pkg/device"
	queue "github.com/perfect-panel/server/queue/types"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdateNodeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNodeLogic {
	return &UpdateNodeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNodeLogic) UpdateNode(req *types.UpdateNodeRequest) error {
	// Check server exist
	nodeInfo, err := l.svcCtx.ServerModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find server error: %v", err)
	}
	tool.DeepCopy(nodeInfo, req, tool.CopyWithIgnoreEmpty(false))
	config, err := json.Marshal(req.Config)
	if err != nil {
		return err
	}

	nodeInfo.Config = string(config)
	nodeRelay, err := json.Marshal(req.RelayNode)
	if err != nil {
		l.Errorw("[UpdateNode] Marshal RelayNode Error: ", logger.Field("error", err.Error()))
		return err
	}

	// 处理Tags字段
	switch {
	case len(req.Tags) > 0:
		// 有Tags，进行连接
		nodeInfo.Tags = strings.Join(req.Tags, ",")
	default:
		// 空数组，清空Tags
		nodeInfo.Tags = ""
	}

	nodeInfo.City = req.City
	nodeInfo.Country = req.Country

	nodeInfo.RelayNode = string(nodeRelay)
	if req.Protocol == "vless" {
		var cfg types.Vless
		if err := json.Unmarshal(config, &cfg); err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "json.Unmarshal error: %v", err.Error())
		}
		if cfg.Security == "reality" && cfg.SecurityConfig.RealityPublicKey == "" {
			public, private, err := tool.Curve25519Genkey(false, "")
			if err != nil {
				return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "generate curve25519 key error")
			}
			cfg.SecurityConfig.RealityPublicKey = public
			cfg.SecurityConfig.RealityPrivateKey = private
			cfg.SecurityConfig.RealityShortId = tool.GenerateShortID(private)
		}
		if cfg.SecurityConfig.RealityServerAddr == "" {
			cfg.SecurityConfig.RealityServerAddr = cfg.SecurityConfig.SNI
		}
		if cfg.SecurityConfig.RealityServerPort == 0 {
			cfg.SecurityConfig.RealityServerPort = 443
		}
		config, _ = json.Marshal(cfg)
		nodeInfo.Config = string(config)
	} else if req.Protocol == "shadowsocks" {
		var cfg types.Shadowsocks
		if err = json.Unmarshal(config, &cfg); err != nil {
			l.Errorf("[CreateNode] Unmarshal Shadowsocks Config Error: %v", err.Error())
			return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "json.Unmarshal error: %v", err.Error())
		}
		if strings.Contains(cfg.Method, "2022") {
			var length int
			switch cfg.Method {
			case "2022-blake3-aes-128-gcm":
				length = 16
			default:
				length = 32
			}
			if len(cfg.ServerKey) != length {
				cfg.ServerKey = tool.GenerateCipher(cfg.ServerKey, length)
			}
		}
		config, _ = json.Marshal(cfg)
		nodeInfo.Config = string(config)
	}
	err = l.svcCtx.ServerModel.Update(l.ctx, nodeInfo)
	if err != nil {
		l.Errorw("[UpdateNode] Update Database Error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create server error: %v", err)
	}

	if req.City == "" || req.Country == "" {
		// Marshal the task payload
		payload, err := json.Marshal(queue.GetNodeCountry{
			Protocol:   nodeInfo.Protocol,
			ServerAddr: nodeInfo.ServerAddr,
		})
		if err != nil {
			l.Errorw("[GetNodeCountry]: Marshal Error", logger.Field("error", err.Error()))
			return errors.Wrap(xerr.NewErrCode(xerr.ERROR), "Failed to marshal task payload")
		}
		// Create a queue task
		task := asynq.NewTask(queue.ForthwithGetCountry, payload)
		// Enqueue the task
		taskInfo, err := l.svcCtx.Queue.Enqueue(task)
		if err != nil {
			l.Errorw("[GetNodeCountry]: Enqueue Error", logger.Field("error", err.Error()), logger.Field("payload", string(payload)))
			return errors.Wrap(xerr.NewErrCode(xerr.ERROR), "Failed to enqueue task")
		}
		l.Infow("[GetNodeCountry]: Enqueue Success", logger.Field("taskID", taskInfo.ID), logger.Field("payload", string(payload)))
	}

	l.svcCtx.DeviceManager.Broadcast(device.SubscribeUpdate)
	return nil
}
