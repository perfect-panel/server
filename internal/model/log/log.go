package log

import (
	"encoding/json"
	"time"
)

type Type uint8

const (
	TypeEmailMessage     Type = iota + 1 // Message log
	TypeMobileMessage                    // Mobile message log
	TypeSubscribe                        // Subscription log
	TypeSubscribeTraffic                 // Subscription traffic log
	TypeServerTraffic                    // Server traffic log
	TypeLogin                            // Login log
	TypeRegister                         // Registration log
	TypeBalance                          // Balance log
	TypeCommission                       // Commission log
	TypeResetSubscribe                   // Reset subscription log
	TypeGift                             // Gift log
)

// Uint8 converts Type to uint8.
func (t Type) Uint8() uint8 {
	return uint8(t)
}

// SystemLog represents a log entry in the system.
type SystemLog struct {
	Id        int64     `gorm:"primaryKey;AUTO_INCREMENT"`
	Type      uint8     `gorm:"index:idx_type;type:tinyint(1);not null;default:0;comment:Log Type: 1: Email Message 2: Mobile Message 3: Subscribe 4: Subscribe Traffic 5: Server Traffic 6: Login 7: Register 8: Balance 9: Commission 10: Reset Subscribe 11: Gift"`
	Date      string    `gorm:"type:varchar(20);default:null;comment:Log Date"`
	ObjectID  int64     `gorm:"index:idx_object_id;type:bigint(20);not null;default:0;comment:Object ID"`
	Content   string    `gorm:"type:text;not null;comment:Log Content"`
	CreatedAt time.Time `gorm:"<-:create;comment:Create Time"`
}

// TableName returns the name of the table for SystemLogs.
func (SystemLog) TableName() string {
	return "system_logs"
}

// Message represents a message log entry.
type Message struct {
	To       string                 `json:"to"`
	Subject  string                 `json:"subject,omitempty"`
	Content  map[string]interface{} `json:"content"`
	Platform string                 `json:"platform"`
	Template string                 `json:"template"`
	Status   uint8                  `json:"status"` // 1: Sent, 2: Failed
}

// Marshal implements the json.Marshaler interface for Message.
func (m *Message) Marshal() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	})
}

// Unmarshal implements the json.Unmarshaler interface for Message.
func (m *Message) Unmarshal(data []byte) error {
	type Alias Message
	aux := (*Alias)(m)
	return json.Unmarshal(data, aux)
}

// Traffic represents a subscription traffic log entry.
type Traffic struct {
	Download int64 `json:"download"`
	Upload   int64 `json:"upload"`
}

// Marshal implements the json.Marshaler interface for SubscribeTraffic.
func (s *Traffic) Marshal() ([]byte, error) {
	type Alias Traffic
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}

// Unmarshal implements the json.Unmarshaler interface for SubscribeTraffic.
func (s *Traffic) Unmarshal(data []byte) error {
	type Alias Traffic
	aux := (*Alias)(s)
	return json.Unmarshal(data, aux)
}

// Login represents a login log entry.
type Login struct {
	LoginIP   string `json:"login_ip"`
	UserAgent string `json:"user_agent"`
	Success   bool   `json:"success"`
	LoginTime int64  `json:"login_time"`
}

// Marshal implements the json.Marshaler interface for Login.
func (l *Login) Marshal() ([]byte, error) {
	type Alias Login
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	})
}

// Unmarshal implements the json.Unmarshaler interface for Login.
func (l *Login) Unmarshal(data []byte) error {
	type Alias Login
	aux := (*Alias)(l)
	return json.Unmarshal(data, aux)
}

// Register represents a registration log entry.
type Register struct {
	AuthMethod   string `json:"auth_method"`
	Identifier   string `json:"identifier"`
	RegisterIP   string `json:"register_ip"`
	UserAgent    string `json:"user_agent"`
	RegisterTime int64  `json:"register_time"`
}

// Marshal implements the json.Marshaler interface for Register.
func (r *Register) Marshal() ([]byte, error) {
	type Alias Register
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// Unmarshal implements the json.Unmarshaler interface for Register.

func (r *Register) Unmarshal(data []byte) error {
	type Alias Register
	aux := (*Alias)(r)
	return json.Unmarshal(data, aux)
}

// Subscribe represents a subscription log entry.
type Subscribe struct {
	Token       string `json:"token"`
	UserAgent   string `json:"user_agent"`
	ClientIP    string `json:"client_ip"`
	SubscribeId int64  `json:"subscribe_id"`
}

// Marshal implements the json.Marshaler interface for Subscribe.
func (s *Subscribe) Marshal() ([]byte, error) {
	type Alias Subscribe
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}

// Unmarshal implements the json.Unmarshaler interface for Subscribe.
func (s *Subscribe) Unmarshal(data []byte) error {
	type Alias Subscribe
	aux := (*Alias)(s)
	return json.Unmarshal(data, aux)
}

const (
	ResetSubscribeTypeAuto    uint8 = 1
	ResetSubscribeTypeAdvance uint8 = 2
	ResetSubscribeTypePaid    uint8 = 3
)

// ResetSubscribe represents a reset subscription log entry.
type ResetSubscribe struct {
	Type    uint8  `json:"type"`
	OrderNo string `json:"order_no,omitempty"`
	ResetAt int64  `json:"reset_at"`
}

// Marshal implements the json.Marshaler interface for ResetSubscribe.
func (r *ResetSubscribe) Marshal() ([]byte, error) {
	type Alias ResetSubscribe
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// Unmarshal implements the json.Unmarshaler interface for ResetSubscribe.
func (r *ResetSubscribe) Unmarshal(data []byte) error {
	type Alias ResetSubscribe
	aux := (*Alias)(r)
	return json.Unmarshal(data, aux)
}

const (
	BalanceTypeRecharge uint8 = 1 // Recharge
	BalanceTypeWithdraw uint8 = 2 // Withdraw
	BalanceTypePayment  uint8 = 3 // Payment
	BalanceTypeRefund   uint8 = 4 // Refund
	BalanceTypeReward   uint8 = 5 // Reward
)

// Balance represents a balance log entry.
type Balance struct {
	Id        int64 `json:"id"`
	Type      uint8 `json:"type"`
	Amount    int64 `json:"amount"`
	OrderId   int64 `json:"order_id,omitempty"`
	Balance   int64 `json:"balance"`
	Timestamp int64 `json:"timestamp"`
}

// Marshal implements the json.Marshaler interface for Balance.
func (b *Balance) Marshal() ([]byte, error) {
	type Alias Balance
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	})
}

// Unmarshal implements the json.Unmarshaler interface for Balance.
func (b *Balance) Unmarshal(data []byte) error {
	type Alias Balance
	aux := (*Alias)(b)
	return json.Unmarshal(data, aux)
}

const (
	CommissionTypePurchase uint8 = 1 // Purchase
	CommissionTypeRenewal  uint8 = 2 // Renewal
	CommissionTypeRefund   uint8 = 3 // Gift
)

// Commission represents a commission log entry.
type Commission struct {
	Type      uint8  `json:"type"`
	Amount    int64  `json:"amount"`
	OrderNo   string `json:"order_no"`
	CreatedAt int64  `json:"created_at"`
}

// Marshal implements the json.Marshaler interface for Commission.
func (c *Commission) Marshal() ([]byte, error) {
	type Alias Commission
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	})
}

// Unmarshal implements the json.Unmarshaler interface for Commission.
func (c *Commission) Unmarshal(data []byte) error {
	type Alias Commission
	aux := (*Alias)(c)
	return json.Unmarshal(data, aux)
}

const (
	GiftTypeIncrease uint8 = 1 // Increase
	GiftTypeReduce   uint8 = 2 // Reduce
)

// Gift represents a gift log entry.
type Gift struct {
	Type        uint8  `json:"type"`
	OrderNo     string `json:"order_no"`
	SubscribeId int64  `json:"subscribe_id"`
	Amount      int64  `json:"amount"`
	Balance     int64  `json:"balance"`
	Remark      string `json:"remark,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// Marshal implements the json.Marshaler interface for Gift.
func (g *Gift) Marshal() ([]byte, error) {
	type Alias Gift
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(g),
	})
}

// Unmarshal implements the json.Unmarshaler interface for Gift.
func (g *Gift) Unmarshal(data []byte) error {
	type Alias Gift
	aux := (*Alias)(g)
	return json.Unmarshal(data, aux)
}
