package telegram

import "fmt"

type AuthData struct {
	Id        *int64  `json:"id,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Username  *string `json:"username,omitempty"`
	PhotoUrl  *string `json:"photo_url,omitempty"`
	AuthDate  *int64  `json:"auth_date,omitempty"`
	Hash      *string `json:"hash,omitempty"`
}

// Validate checks the hash of AuthData with computed one. To compute hash botToken is required.
// Ref: https://core.telegram.org/widgets/login#checking-authorization
func (d *AuthData) Validate(botToken []byte) error {
	if d.Hash == nil {
		return fmt.Errorf("auth data has no 'hash' value")
	}
	if len(botToken) == 0 {
		return fmt.Errorf("telegram bot token is not provided")
	}
	hash := *d.Hash
	computedHash := computeHash(d, botToken)
	if hash != computedHash {
		return fmt.Errorf("hash is not valid")
	}
	return nil
}
