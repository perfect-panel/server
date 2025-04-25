package auth

import (
	"context"

	"github.com/perfect-panel/server/pkg/authmethod"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/phone"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func findUserByMethod(ctx context.Context, svcCtx *svc.ServiceContext, method, identifier, account, areaCode string) (userInfo *user.User, err error) {
	var authMethods *user.AuthMethods
	switch method {
	case authmethod.Email:
		authMethods, err = svcCtx.UserModel.FindUserAuthMethodByOpenID(ctx, authmethod.Email, account)
	case authmethod.Mobile:
		phoneNumber, err := phone.FormatToE164(areaCode, account)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.TelephoneError), "Invalid phone number")
		}
		authMethods, err = svcCtx.UserModel.FindUserAuthMethodByOpenID(ctx, authmethod.Mobile, phoneNumber)
		if err != nil {
			return nil, err
		}
	case authmethod.Device:
		userDevice, err := svcCtx.UserModel.FindOneDeviceByIdentifier(ctx, identifier)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query user device imei error")
		}
		return svcCtx.UserModel.FindOne(ctx, userDevice.UserId)
	default:
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.UserNotExist), "unknown method")
	}
	if err != nil {
		return nil, err
	}
	return svcCtx.UserModel.FindOne(ctx, authMethods.UserId)
}

func existError(method string) error {
	switch method {
	case authmethod.Email:
		return errors.Wrapf(xerr.NewErrCode(xerr.EmailExist), "")
	case authmethod.Mobile:
		return errors.Wrapf(xerr.NewErrCode(xerr.TelephoneExist), "")
	case authmethod.Device:
		return errors.Wrapf(xerr.NewErrCode(xerr.DeviceExist), "")
	default:
		return errors.New("unknown method")
	}
}
