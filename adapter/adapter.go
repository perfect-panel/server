package adapter

import (
	"encoding/json"

	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/random"
)

type Adapter struct {
	SiteName       string           // 站点名称
	Servers        []*server.Server // 服务器列表
	UserInfo       User             // 用户信息
	ClientTemplate string           // 客户端配置模板
	OutputFormat   string           // 输出格式，默认是 base64
	SubscribeName  string           // 订阅名称
}

type Option func(*Adapter)

// WithServers 设置服务器列表
func WithServers(servers []*server.Server) Option {
	return func(opts *Adapter) {
		opts.Servers = servers
	}
}

// WithUserInfo 设置用户信息
func WithUserInfo(user User) Option {
	return func(opts *Adapter) {
		opts.UserInfo = user
	}
}

// WithOutputFormat 设置输出格式
func WithOutputFormat(format string) Option {
	return func(opts *Adapter) {
		opts.OutputFormat = format
	}
}

// WithSiteName 设置站点名称
func WithSiteName(name string) Option {
	return func(opts *Adapter) {
		opts.SiteName = name
	}
}

// WithSubscribeName 设置订阅名称
func WithSubscribeName(name string) Option {
	return func(opts *Adapter) {
		opts.SubscribeName = name
	}
}

func NewAdapter(tpl string, opts ...Option) *Adapter {
	adapter := &Adapter{
		Servers:        []*server.Server{},
		UserInfo:       User{},
		ClientTemplate: tpl,
		OutputFormat:   "base64", // 默认输出格式
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

func (adapter *Adapter) Client() (*Client, error) {
	client := &Client{
		SiteName:       adapter.SiteName,
		SubscribeName:  adapter.SubscribeName,
		ClientTemplate: adapter.ClientTemplate,
		OutputFormat:   adapter.OutputFormat,
		Proxies:        []Proxy{},
		UserInfo:       adapter.UserInfo,
	}

	proxies, err := adapter.Proxies(adapter.Servers)
	if err != nil {
		return nil, err
	}
	client.Proxies = proxies
	return client, nil
}

func (adapter *Adapter) Proxies(servers []*server.Server) ([]Proxy, error) {
	var proxies []Proxy
	for _, srv := range servers {
		switch srv.RelayMode {
		case server.RelayModeAll:
			var relays []server.NodeRelay
			if err := json.Unmarshal([]byte(srv.RelayNode), &relays); err != nil {
				logger.Errorw("Unmarshal RelayNode", logger.Field("error", err.Error()), logger.Field("node", srv.Name), logger.Field("relayNode", srv.RelayNode))
				continue
			}
			for _, relay := range relays {
				proxy, err := adapterProxy(*srv, relay.Host, uint64(relay.Port))
				if err != nil {
					logger.Errorw("Adapter Proxy", logger.Field("error", err.Error()), logger.Field("node", srv.Name), logger.Field("relayNode", relay))
					continue
				}
				proxies = append(proxies, proxy)
			}

		case server.RelayModeRandom:
			var relays []server.NodeRelay
			if err := json.Unmarshal([]byte(srv.RelayNode), &relays); err != nil {
				logger.Errorw("Unmarshal RelayNode", logger.Field("error", err.Error()), logger.Field("node", srv.Name), logger.Field("relayNode", srv.RelayNode))
				continue
			}
			randNum := random.RandomInRange(0, len(relays)-1)
			relay := relays[randNum]
			proxy, err := adapterProxy(*srv, relay.Host, uint64(relay.Port))
			if err != nil {
				logger.Errorw("Adapter Proxy", logger.Field("error", err.Error()), logger.Field("node", srv.Name), logger.Field("relayNode", relay))
				continue
			}
			proxies = append(proxies, proxy)

		case server.RelayModeNone:
			proxy, err := adapterProxy(*srv, srv.ServerAddr, 0)
			if err != nil {
				logger.Errorw("Adapter Proxy", logger.Field("error", err.Error()), logger.Field("node", srv.Name), logger.Field("serverAddr", srv.ServerAddr))
				continue
			}
			proxies = append(proxies, proxy)
		default:
			logger.Errorw("Unknown RelayMode", logger.Field("node", srv.Name), logger.Field("relayMode", srv.RelayMode))
		}

	}
	return proxies, nil
}
