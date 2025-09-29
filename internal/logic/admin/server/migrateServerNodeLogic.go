package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type MigrateServerNodeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewMigrateServerNodeLogic Migrate server and node data to new database
func NewMigrateServerNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MigrateServerNodeLogic {
	return &MigrateServerNodeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MigrateServerNodeLogic) MigrateServerNode() (resp *types.MigrateServerNodeResponse, err error) {
	tx := l.svcCtx.DB.WithContext(l.ctx).Begin()
	var oldServers []*server.Server
	var newServers []*node.Server
	var newNodes []*node.Node

	err = tx.Model(&server.Server{}).Find(&oldServers).Error
	if err != nil {
		l.Errorw("[MigrateServerNode] Query Old Server List Error: ", logger.Field("error", err.Error()))
		return &types.MigrateServerNodeResponse{
			Succee:  0,
			Fail:    0,
			Message: fmt.Sprintf("Query Old Server List Error: %s", err.Error()),
		}, nil
	}
	for _, oldServer := range oldServers {
		data, err := l.adapterServer(oldServer)
		if err != nil {
			l.Errorw("[MigrateServerNode] Adapter Server Error: ", logger.Field("error", err.Error()))
			if resp == nil {
				resp = &types.MigrateServerNodeResponse{}
			}
			resp.Fail++
			if resp.Message == "" {
				resp.Message = fmt.Sprintf("Adapter Server Error: %s", err.Error())
			} else {
				resp.Message = fmt.Sprintf("%s; Adapter Server Error: %s", resp.Message, err.Error())
			}
			continue
		}
		newServers = append(newServers, data)

		newNode, err := l.adapterNode(oldServer)
		if err != nil {
			l.Errorw("[MigrateServerNode] Adapter Node Error: ", logger.Field("error", err.Error()))
			if resp == nil {
				resp = &types.MigrateServerNodeResponse{}
			}
			resp.Fail++
			if resp.Message == "" {
				resp.Message = fmt.Sprintf("Adapter Node Error: %s", err.Error())
			} else {
				resp.Message = fmt.Sprintf("%s; Adapter Node Error: %s", resp.Message, err.Error())
			}
			continue
		}
		for _, item := range newNode {
			if item.Port == 0 {
				protocols, _ := data.UnmarshalProtocols()
				if len(protocols) > 0 {
					item.Port = protocols[0].Port
				}
			}
			newNodes = append(newNodes, item)
		}
	}

	if len(newServers) > 0 {
		err = tx.Model(&node.Server{}).CreateInBatches(newServers, 20).Error
		if err != nil {
			tx.Rollback()
			l.Errorw("[MigrateServerNode] Insert New Server List Error: ", logger.Field("error", err.Error()))
			return &types.MigrateServerNodeResponse{
				Succee:  0,
				Fail:    uint64(len(newServers)),
				Message: fmt.Sprintf("Insert New Server List Error: %s", err.Error()),
			}, nil
		}
	}
	if len(newNodes) > 0 {
		err = tx.Model(&node.Node{}).CreateInBatches(newNodes, 20).Error
		if err != nil {
			tx.Rollback()
			l.Errorw("[MigrateServerNode] Insert New Node List Error: ", logger.Field("error", err.Error()))
			return &types.MigrateServerNodeResponse{
				Succee:  uint64(len(newServers)),
				Fail:    uint64(len(newNodes)),
				Message: fmt.Sprintf("Insert New Node List Error: %s", err.Error()),
			}, nil
		}
	}
	tx.Commit()

	return &types.MigrateServerNodeResponse{
		Succee:  uint64(len(newServers)),
		Fail:    0,
		Message: fmt.Sprintf("Migrate Success: %d servers and %d nodes", len(newServers), len(newNodes)),
	}, nil
}

func (l *MigrateServerNodeLogic) adapterServer(info *server.Server) (*node.Server, error) {
	result := &node.Server{
		Id:      info.Id,
		Name:    info.Name,
		Country: info.Country,
		City:    info.City,
		//Ratio:     info.TrafficRatio,
		Address:   info.ServerAddr,
		Sort:      int(info.Sort),
		Protocols: "",
	}
	var protocols []node.Protocol

	switch info.Protocol {
	case ShadowSocks:
		var src server.Shadowsocks
		err := json.Unmarshal([]byte(info.Config), &src)
		if err != nil {
			return nil, err
		}
		protocols = append(protocols, node.Protocol{
			Type:      "shadowsocks",
			Cipher:    src.Method,
			Port:      uint16(src.Port),
			ServerKey: src.ServerKey,
			Ratio:     float64(info.TrafficRatio),
		})
	case Vmess:
		var src server.Vmess
		err := json.Unmarshal([]byte(info.Config), &src)
		if err != nil {
			return nil, err
		}
		protocol := node.Protocol{
			Type:              "vmess",
			Port:              uint16(src.Port),
			Security:          src.Security,
			SNI:               src.SecurityConfig.SNI,
			AllowInsecure:     src.SecurityConfig.AllowInsecure,
			Fingerprint:       src.SecurityConfig.Fingerprint,
			RealityServerAddr: src.SecurityConfig.RealityServerAddr,
			RealityServerPort: src.SecurityConfig.RealityServerPort,
			RealityPrivateKey: src.SecurityConfig.RealityPrivateKey,
			RealityPublicKey:  src.SecurityConfig.RealityPublicKey,
			RealityShortId:    src.SecurityConfig.RealityShortId,
			Transport:         src.Transport,
			Host:              src.TransportConfig.Host,
			Path:              src.TransportConfig.Path,
			ServiceName:       src.TransportConfig.ServiceName,
			Flow:              src.Flow,
			Ratio:             float64(info.TrafficRatio),
		}
		protocols = append(protocols, protocol)
		protocols = append(protocols, protocol)
	case Vless:
		var src server.Vless
		err := json.Unmarshal([]byte(info.Config), &src)
		if err != nil {
			return nil, err
		}
		protocol := node.Protocol{
			Type:              "vless",
			Port:              uint16(src.Port),
			Security:          src.Security,
			SNI:               src.SecurityConfig.SNI,
			AllowInsecure:     src.SecurityConfig.AllowInsecure,
			Fingerprint:       src.SecurityConfig.Fingerprint,
			RealityServerAddr: src.SecurityConfig.RealityServerAddr,
			RealityServerPort: src.SecurityConfig.RealityServerPort,
			RealityPrivateKey: src.SecurityConfig.RealityPrivateKey,
			RealityPublicKey:  src.SecurityConfig.RealityPublicKey,
			RealityShortId:    src.SecurityConfig.RealityShortId,
			Transport:         src.Transport,
			Host:              src.TransportConfig.Host,
			Path:              src.TransportConfig.Path,
			ServiceName:       src.TransportConfig.ServiceName,
			Flow:              src.Flow,
			Ratio:             float64(info.TrafficRatio),
		}
		protocols = append(protocols, protocol)
	case Trojan:
		var src server.Trojan
		err := json.Unmarshal([]byte(info.Config), &src)
		if err != nil {
			return nil, err
		}
		protocol := node.Protocol{
			Type:              "trojan",
			Port:              uint16(src.Port),
			Security:          src.Security,
			SNI:               src.SecurityConfig.SNI,
			AllowInsecure:     src.SecurityConfig.AllowInsecure,
			Fingerprint:       src.SecurityConfig.Fingerprint,
			RealityServerAddr: src.SecurityConfig.RealityServerAddr,
			RealityServerPort: src.SecurityConfig.RealityServerPort,
			RealityPrivateKey: src.SecurityConfig.RealityPrivateKey,
			RealityPublicKey:  src.SecurityConfig.RealityPublicKey,
			RealityShortId:    src.SecurityConfig.RealityShortId,
			Transport:         src.Transport,
			Host:              src.TransportConfig.Host,
			Path:              src.TransportConfig.Path,
			ServiceName:       src.TransportConfig.ServiceName,
			Flow:              src.Flow,
			Ratio:             float64(info.TrafficRatio),
		}
		protocols = append(protocols, protocol)
	case Hysteria2:
		var src server.Hysteria2
		err := json.Unmarshal([]byte(info.Config), &src)
		if err != nil {
			return nil, err
		}
		protocol := node.Protocol{
			Type:              "hysteria",
			Port:              uint16(src.Port),
			HopPorts:          src.HopPorts,
			HopInterval:       src.HopInterval,
			ObfsPassword:      src.ObfsPassword,
			SNI:               src.SecurityConfig.SNI,
			AllowInsecure:     src.SecurityConfig.AllowInsecure,
			Fingerprint:       src.SecurityConfig.Fingerprint,
			RealityServerAddr: src.SecurityConfig.RealityServerAddr,
			RealityServerPort: src.SecurityConfig.RealityServerPort,
			RealityPrivateKey: src.SecurityConfig.RealityPrivateKey,
			RealityPublicKey:  src.SecurityConfig.RealityPublicKey,
			RealityShortId:    src.SecurityConfig.RealityShortId,
			Ratio:             float64(info.TrafficRatio),
		}
		protocols = append(protocols, protocol)
	case Tuic:
		var src server.Tuic
		err := json.Unmarshal([]byte(info.Config), &src)
		if err != nil {
			return nil, err
		}
		protocol := node.Protocol{
			Type:                 "tuic",
			Port:                 uint16(src.Port),
			DisableSNI:           src.DisableSNI,
			ReduceRtt:            src.ReduceRtt,
			UDPRelayMode:         src.UDPRelayMode,
			CongestionController: src.CongestionController,
			SNI:                  src.SecurityConfig.SNI,
			AllowInsecure:        src.SecurityConfig.AllowInsecure,
			Fingerprint:          src.SecurityConfig.Fingerprint,
			RealityServerAddr:    src.SecurityConfig.RealityServerAddr,
			RealityServerPort:    src.SecurityConfig.RealityServerPort,
			RealityPrivateKey:    src.SecurityConfig.RealityPrivateKey,
			RealityPublicKey:     src.SecurityConfig.RealityPublicKey,
			RealityShortId:       src.SecurityConfig.RealityShortId,
			Ratio:                float64(info.TrafficRatio),
		}
		protocols = append(protocols, protocol)
	case AnyTLS:
		var src server.AnyTLS
		err := json.Unmarshal([]byte(info.Config), &src)
		if err != nil {
			return nil, err
		}
		protocol := node.Protocol{
			Type:              "anytls",
			Port:              uint16(src.Port),
			SNI:               src.SecurityConfig.SNI,
			AllowInsecure:     src.SecurityConfig.AllowInsecure,
			Fingerprint:       src.SecurityConfig.Fingerprint,
			RealityServerAddr: src.SecurityConfig.RealityServerAddr,
			RealityServerPort: src.SecurityConfig.RealityServerPort,
			RealityPrivateKey: src.SecurityConfig.RealityPrivateKey,
			RealityPublicKey:  src.SecurityConfig.RealityPublicKey,
			RealityShortId:    src.SecurityConfig.RealityShortId,
			Ratio:             float64(info.TrafficRatio),
		}
		protocols = append(protocols, protocol)
	}
	if len(protocols) > 0 {
		err := result.MarshalProtocols(protocols)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (l *MigrateServerNodeLogic) adapterNode(info *server.Server) ([]*node.Node, error) {
	var nodes []*node.Node
	enable := true
	switch info.RelayMode {
	case server.RelayModeNone:
		nodes = append(nodes, &node.Node{
			Name:     info.Name,
			Tags:     "",
			Port:     0,
			Address:  info.ServerAddr,
			ServerId: info.Id,
			Protocol: info.Protocol,
			Enabled:  &enable,
		})
	default:
		var relays []server.NodeRelay
		err := json.Unmarshal([]byte(info.RelayNode), &relays)
		if err != nil {
			return nil, err
		}
		for _, relay := range relays {
			nodes = append(nodes, &node.Node{
				Name:     relay.Prefix + info.Name,
				Tags:     "",
				Port:     uint16(relay.Port),
				Address:  relay.Host,
				ServerId: info.Id,
				Protocol: info.Protocol,
				Enabled:  &enable,
			})
		}
	}

	return nodes, nil
}
