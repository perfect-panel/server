package ws

import (
	"context"
	sysErr "errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

type DeviceWsConnectLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Webosocket Device Connect
func NewDeviceWsConnectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeviceWsConnectLogic {
	return &DeviceWsConnectLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeviceWsConnectLogic) DeviceWsConnect(c *gin.Context) error {

	value := l.ctx.Value(constant.CtxKeyIdentifier)
	if value == nil || value.(string) == "" {
		l.Errorf("DeviceWsConnectLogic DeviceWsConnect identifier is empty")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "identifier is empty")
	}
	identifier := value.(string)
	_, err := l.svcCtx.UserModel.FindOneDeviceByIdentifier(l.ctx, identifier)
	if err != nil && !sysErr.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf("DeviceWsConnectLogic DeviceWsConnect FindOneDeviceByIdentifier err: %v", err)
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), err.Error())
	}

	value = l.ctx.Value(constant.CtxKeyUser)
	if value == nil {
		l.Errorf("DeviceWsConnectLogic DeviceWsConnect value is nil")
		return nil
	}
	userInfo := value.(*user.User)
	if sysErr.Is(err, gorm.ErrRecordNotFound) {
		device := user.Device{
			Identifier: identifier,
			UserId:     userInfo.Id,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			Online:     true,
			Enabled:    true,
		}
		err := l.svcCtx.UserModel.InsertDevice(l.ctx, &device)
		if err != nil {
			l.Errorf("DeviceWsConnectLogic DeviceWsConnect InsertDevice err: %v", err)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), err.Error())
		}
	}
	//默认在线设备1
	maxDevice := 3
	subscribe, err := l.svcCtx.UserModel.QueryUserSubscribe(l.ctx, userInfo.Id, 1, 2)
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
	l.svcCtx.DeviceManager.AddDevice(c.Writer, c.Request, l.ctx.Value(constant.CtxKeySessionID).(string), userInfo.Id, identifier, maxDevice)
	return nil
}
