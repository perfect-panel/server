package xerr

/** (The first 3 digits represent the business, and the last three digits represent the specific function) **/

// General error code
const (
	SUCCESS uint32 = 200
	ERROR   uint32 = 500
)

// Database error
const (
	DatabaseQueryError   uint32 = 10001
	DatabaseUpdateError  uint32 = 10002
	DatabaseInsertError  uint32 = 10003
	DatabaseDeletedError uint32 = 10004
)

// User error
const (
	UserExist           uint32 = 20001
	UserNotExist        uint32 = 20002
	UserPasswordError   uint32 = 20003
	UserDisabled        uint32 = 20004
	InsufficientBalance uint32 = 20005
	StopRegister        uint32 = 20006
	TelegramNotBound    uint32 = 20007
	UserNotBindOauth    uint32 = 20008
	InviteCodeError     uint32 = 20009
)

// Node error
const (
	NodeExist         uint32 = 30001
	NodeNotExist      uint32 = 30002
	NodeGroupExist    uint32 = 30003
	NodeGroupNotExist uint32 = 30004
	NodeGroupNotEmpty uint32 = 30005
)

// Request error
const (
	InvalidParams     uint32 = 400
	TooManyRequests   uint32 = 401
	ErrorTokenEmpty   uint32 = 40002
	ErrorTokenInvalid uint32 = 40003
	ErrorTokenExpire  uint32 = 40004
	InvalidAccess     uint32 = 40005
	InvalidCiphertext uint32 = 40006
	SecretIsEmpty     uint32 = 40007
)

//coupon error

const (
	CouponNotExist          uint32 = 50001 // Coupon does not exist
	CouponAlreadyUsed       uint32 = 50002 // Coupon has already been used
	CouponNotApplicable     uint32 = 50003 // Coupon does not match the order or conditions
	CouponInsufficientUsage uint32 = 50004 // Coupon has insufficient remaining uses
	CouponExpired           uint32 = 50005 // Coupon is expired
)

// Subscribe

const (
	SubscribeExpired                uint32 = 60001
	SubscribeNotAvailable           uint32 = 60002
	UserSubscribeExist              uint32 = 60003
	SubscribeIsUsedError            uint32 = 60004
	SingleSubscribeModeExceedsLimit uint32 = 60005
	SubscribeQuotaLimit             uint32 = 60006
)

// Auth error

const (
	VerifyCodeError uint32 = 70001
)

// equipment error

const (
	QueueEnqueueError uint32 = 80001
)

// System error

const (
	DebugModeError uint32 = 90001
)

const (
	SendSmsError    uint32 = 90002
	SmsNotEnabled   uint32 = 90003
	EmailNotEnabled uint32 = 90004
)

const (
	GetAuthenticatorError          uint32 = 90005
	AuthenticatorNotSupportedError uint32 = 90006
	TelephoneAreaCodeIsEmpty       uint32 = 90007
	TodaySendCountExceedsLimit     uint32 = 90015
)

const (
	PasswordIsEmpty                    uint32 = 90008
	AreaCodeIsEmpty                    uint32 = 90009
	PasswordOrVerificationCodeRequired uint32 = 90010
	EmailExist                         uint32 = 90011
	TelephoneExist                     uint32 = 90012
	DeviceExist                        uint32 = 90013
	TelephoneError                     uint32 = 90014
)
const (
	DeviceNotExist uint32 = 90017
	UseridNotMatch uint32 = 90018
)

const (
	OrderNotExist         uint32 = 61001
	PaymentMethodNotFound uint32 = 61002
	OrderStatusError      uint32 = 61003
	InsufficientOfPeriod  uint32 = 61004
	ExistAvailableTraffic uint32 = 61005
)
