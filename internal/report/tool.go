package report

import (
	"fmt"
	"net"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/pkg/errors"
)

// FreePort returns a free TCP port by opening a listener on port 0.
func FreePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	// Get the assigned port
	addr := l.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// ModulePort returns the module port from the environment variable or a free port.
func ModulePort() (int, error) {
	// 从环境变量获取端口号
	value, exists := os.LookupEnv("PPANEL_PORT")
	if exists {
		var port int
		_, err := fmt.Sscanf(value, "%d", &port)
		if err != nil {
			return FreePort()
		}
		return port, nil
	}
	return FreePort()
}

// GatewayPort returns the gateway port from the environment variable or a free port.
func GatewayPort() (int, error) {
	// 从环境变量获取端口号
	value, exists := os.LookupEnv("GATEWAY_PORT")
	if exists {
		var port int
		_, err := fmt.Sscanf(value, "%d", &port)
		if err != nil {
			logger.Errorf("Failed to parse GATEWAY_PORT: %v Value %s", err.Error(), value)
			panic(err)
		}
		return port, nil
	}
	return 0, errors.New("could not determine gateway port")
}

// RegisterModule registers a module with the gateway.
func RegisterModule(port int) error {
	// 从环境变量中读取网关模块端口
	gatewayPort, err := GatewayPort()
	if err != nil {
		return err
	}

	// 从环境变量中获取通讯密钥
	value, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		panic("could not determine secret key")
	}

	var response RegisterResponse

	client := resty.New().SetBaseURL(fmt.Sprintf("http://127.0.0.1:%d", gatewayPort))
	result, err := client.R().SetHeader("Content-Type", "application/json").SetBody(RegisterServiceRequest{
		Secret:         value,
		ProxyPath:      "/api",
		ServiceURL:     fmt.Sprintf("http://127.0.0.1:%d", port),
		Repository:     constant.Repository,
		HeartbeatURL:   fmt.Sprintf("http://127.0.0.1:%d/v1/common/heartbeat", port),
		ServiceName:    constant.ServiceName,
		ServiceVersion: constant.Version,
	}).SetResult(&response).Post(RegisterAPI)

	if err != nil {
		return err
	}

	if result.IsError() {
		return errors.New("failed to register module: " + result.Status())
	}

	if !response.Success {
		return errors.New("failed to register module: " + response.Message)
	}
	logger.Infof("Module registered successfully: %s", response.Message)
	return nil
}

// IsGatewayMode checks if the application is running in gateway mode.
// It returns true if GATEWAY_MODE is set to "true" and GATEWAY_PORT is valid.
func IsGatewayMode() bool {
	value, exists := os.LookupEnv("GATEWAY_MODE")
	if exists && value == "true" {
		if _, err := GatewayPort(); err == nil {
			return true
		}
	}

	return false
}
