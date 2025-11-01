package report

import (
	"fmt"
	"net"
	"os"

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
			panic(err)
		}
		return port, nil
	}
	return 0, errors.New("could not determine gateway port")
}
