package server

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
)

type QueryServerProtocolConfigLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewQueryServerProtocolConfigLogic Get Server Protocol Config
func NewQueryServerProtocolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryServerProtocolConfigLogic {
	return &QueryServerProtocolConfigLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryServerProtocolConfigLogic) QueryServerProtocolConfig(req *types.QueryServerConfigRequest) (resp *types.QueryServerConfigResponse, err error) {
	// find server
	data, err := l.svcCtx.NodeModel.FindOneServer(l.ctx, req.ServerID)
	if err != nil {
		l.Errorf("[GetServerProtocols] FindOneServer Error: %s", err.Error())
		return nil, err
	}

	// handler protocols
	var protocols []types.Protocol
	dst, err := data.UnmarshalProtocols()
	if err != nil {
		l.Errorf("[FilterServerList] UnmarshalProtocols Error: %s", err.Error())
		return nil, err
	}
	tool.DeepCopy(&protocols, dst)

	// filter by req.Protocols

	if len(req.Protocols) > 0 {
		var filtered []types.Protocol
		protocolSet := make(map[string]struct{})
		for _, p := range req.Protocols {
			protocolSet[p] = struct{}{}
		}
		for _, p := range protocols {
			if _, exists := protocolSet[p.Type]; exists {
				filtered = append(filtered, p)
			}
		}
		protocols = filtered
	}

	var dns []types.NodeDNS
	if len(l.svcCtx.Config.Node.DNS) > 0 {
		for _, d := range l.svcCtx.Config.Node.DNS {
			dns = append(dns, types.NodeDNS{
				Proto:   d.Proto,
				Address: d.Address,
				Domains: d.Domains,
			})
		}
	}
	var outbound []types.NodeOutbound
	if len(l.svcCtx.Config.Node.Outbound) > 0 {
		for _, o := range l.svcCtx.Config.Node.Outbound {
			outbound = append(outbound, types.NodeOutbound{
				Name:     o.Name,
				Protocol: o.Protocol,
				Address:  o.Address,
				Port:     o.Port,
				Password: o.Password,
				Rules:    o.Rules,
			})
		}
	}

	return &types.QueryServerConfigResponse{
		TrafficReportThreshold: l.svcCtx.Config.Node.TrafficReportThreshold,
		IPStrategy:             l.svcCtx.Config.Node.IPStrategy,
		DNS:                    dns,
		Block:                  l.svcCtx.Config.Node.Block,
		Outbound:               outbound,
		Protocols:              protocols,
		Total:                  int64(len(protocols)),
	}, nil
}
