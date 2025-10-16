package xerr

var message map[uint32]string

func init() {
	message = make(map[uint32]string)
	message = map[uint32]string{
		// General error
		SUCCESS: "Success",
		ERROR:   "Internal Server Error",
		// parameter error
		TooManyRequests:   "Too Many Requests",
		InvalidParams:     "Param Error",
		ErrorTokenEmpty:   "User token is empty",
		ErrorTokenInvalid: "User token is invalid",
		ErrorTokenExpire:  "User token is expired",
		SecretIsEmpty:     "Secret is empty",
		InvalidAccess:     "Invalid access",
		InvalidCiphertext: "Invalid ciphertext",
		// Database error
		DatabaseQueryError:   "Database query error",
		DatabaseUpdateError:  "Database update error",
		DatabaseInsertError:  "Database insert error",
		DatabaseDeletedError: "Database deleted error",

		// User error
		UserExist:           "User already exists",
		UserNotExist:        "User does not exist",
		UserPasswordError:   "User password error",
		UserDisabled:        "User disabled",
		InsufficientBalance: "Insufficient balance",
		StopRegister:        "Stop register",
		TelegramNotBound:    "Telegram not bound ",
		UserNotBindOauth:    "User not bind oauth method",
		InviteCodeError:     "Invite code error",

		// Node error
		NodeExist:         "Node already exists",
		NodeNotExist:      "Node does not exist",
		NodeGroupExist:    "Node group already exists",
		NodeGroupNotExist: "Node group does not exist",
		NodeGroupNotEmpty: "Node group is not empty",

		//coupon error
		CouponNotExist:          "Coupon does not exist",
		CouponAlreadyUsed:       "Coupon has already been used",
		CouponNotApplicable:     "Coupon does not match the order or conditions",
		CouponInsufficientUsage: "Coupon has insufficient remaining uses",
		CouponExpired:           "Coupon is expired",

		// Subscribe
		SubscribeExpired:                "Subscribe is expired",
		SubscribeNotAvailable:           "Subscribe is not available",
		UserSubscribeExist:              "User has subscription",
		SubscribeIsUsedError:            "Subscribe is used",
		SingleSubscribeModeExceedsLimit: "Single subscribe mode exceeds limit",
		SubscribeQuotaLimit:             "Subscribe quota limit",

		// auth error
		VerifyCodeError: "Verify code error",

		// EnqueueError
		QueueEnqueueError: " Queue enqueue error",

		// System error
		DebugModeError: "Debug mode is enabled",

		GetAuthenticatorError:          "Unsupported login method",
		AuthenticatorNotSupportedError: "The authenticator does not support this method",

		TelephoneAreaCodeIsEmpty:           "Telephone area code is empty",
		TodaySendCountExceedsLimit:         "This account has reached the limit of sending times today",
		SmsNotEnabled:                      "Telephone login is not enabled",
		EmailNotEnabled:                    "Email function is not enabled yet",
		PasswordOrVerificationCodeRequired: "Password or verification code required",
		EmailExist:                         "Email already exists",
		TelephoneExist:                     "Telephone already exists",
		DeviceExist:                        "device exists",
		PasswordIsEmpty:                    "password is empty",
		TelephoneError:                     "telephone number error",
		DeviceNotExist:                     "Device does not exist",
		UseridNotMatch:                     "Userid not match",

		// Order error
		OrderNotExist:         "Order does not exist",
		PaymentMethodNotFound: "Payment method not found",
		OrderStatusError:      "Order status error",
		InsufficientOfPeriod:  "Insufficient number of period",
	}

}

func MapErrMsg(errCode uint32) string {
	if msg, ok := message[errCode]; ok {
		return msg
	} else {
		return "Internal Server Error"
	}
}

func IsCodeErr(errCode uint32) bool {
	if _, ok := message[errCode]; ok {
		return true
	} else {
		return false
	}
}
