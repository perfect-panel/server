package clash

import "github.com/perfect-panel/server/pkg/adapter/proxy"

func clashTransport(c *Proxy, transportType string, transportConfig proxy.TransportConfig) {

	switch transportType {
	case "websocket", "httpupgrade":
		if transportType == "websocket" {
			c.Network = "ws"
		} else {
			c.Network = transportType
		}
		c.WSOpts = WSOptions{
			Path:    transportConfig.Path,
			Headers: map[string]string{},
		}
		if transportConfig.Host != "" {
			c.WSOpts.Headers["host"] = transportConfig.Host
		}
		if transportType == "httpupgrade" {
			c.WSOpts.V2rayHttpUpgrade = true
		}
	case "grpc":
		c.Network = "grpc"
		c.GrpcOpts = GrpcOptions{
			GrpcServiceName: transportConfig.ServiceName,
		}
	case "tcp":
		c.Network = "tcp"
	}

}
