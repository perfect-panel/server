package ws

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/pkg/constant"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
)

type AppWsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// App heartbeat
func NewAppWsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppWsLogic {
	return &AppWsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppWsLogic) AppWs(w http.ResponseWriter, r *http.Request, userid, identifier string) error {
	//获取设备号
	if identifier == "" {
		return xerr.NewErrCode(xerr.DeviceNotExist)
	}
	//获取用户id
	userID, err := strconv.ParseInt(userid, 10, 64)
	if err != nil {
		return xerr.NewErrCode(xerr.UseridNotMatch)
	}

	////获取session
	value := l.ctx.Value(constant.CtxKeySessionID)
	if value == nil {
		return xerr.NewErrCode(xerr.ErrorTokenInvalid)
	}
	session := value.(string)

	//获取用户
	userInfo := l.ctx.Value(constant.CtxKeyUser).(*user.User)

	if userID != userInfo.Id {
		return xerr.NewErrCode(xerr.UseridNotMatch)
	}

	_, err = l.svcCtx.UserModel.FindOneDeviceByIdentifier(l.ctx, identifier)
	if err != nil {
		return xerr.NewErrCode(xerr.DeviceNotExist)
	}

	//if device.UserId != userInfo.Id {
	//	return xerr.NewErrCode(xerr.DeviceNotExist)
	//}

	//默认在线设备1
	maxDevice := 0
	subscribe, err := l.svcCtx.UserModel.QueryUserSubscribe(l.ctx, userInfo.Id)
	if err == nil {
		for _, sub := range subscribe {
			if time.Now().Before(sub.ExpireTime) {
				deviceLimit := int(sub.Subscribe.DeviceLimit)
				if deviceLimit > maxDevice {
					maxDevice = deviceLimit
				}
			}
		}
	}
	l.svcCtx.DeviceManager.AddDevice(w, r, session, userID, identifier, maxDevice)
	return nil
}
