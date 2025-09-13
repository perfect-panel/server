package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/result"
)

// QueryServerProtocolConfigHandler Get Server Protocol Config
func QueryServerProtocolConfigHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.QueryServerConfigRequest

		serverID, err := strconv.ParseInt(c.Param("server_id"), 10, 64)
		if err != nil {
			logger.Debugf("[QueryServerProtocolConfigHandler] - strconv.ParseInt(server_id) error: %v, Param: %s", err, c.Param("server_id"))
			c.String(http.StatusBadRequest, "Invalid Params")
			c.Abort()
			return
		}
		req.ServerID = serverID

		key := c.GetHeader("secret_key")
		if key == "" || key != svcCtx.Config.Node.NodeSecret {
			logger.Debugf("[QueryServerProtocolConfigHandler] - secret_key error: %s", key)
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		l := server.NewQueryServerProtocolConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryServerProtocolConfig(&req)
		result.HttpResult(c, resp, err)
	}
}
