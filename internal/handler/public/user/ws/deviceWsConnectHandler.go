package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	wslogic "github.com/perfect-panel/server/internal/logic/public/user/ws"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境中应该根据需求限制
	},
}

// Webosocket Device Connect
func DeviceWsConnectHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := wslogic.NewDeviceWsConnectLogic(c.Request.Context(), svcCtx)
		err := l.DeviceWsConnect(c)
		result.HttpResult(c, nil, err)
	}
}
