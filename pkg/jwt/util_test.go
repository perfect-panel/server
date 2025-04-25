package jwt

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

// TestNewJwtToken test NewJwtToken function
func TestParseJwtToken(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJEZXZpY2VJZCI6IjM4IiwiZXhwIjoxNzE4MTU2OTQ4LCJpYXQiOjE3MTc1NTIxNDgsInVzZXJJZCI6MX0.4W0nga82kNrfwWjkwcgYAWj4fI4iRc-ZftwVbu-a_kI"
	secret := "ae0536f9-6450-4606-8e13-5a19ed505da0"

	claims, err := ParseJwtToken(token, secret)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		t.Errorf("err: %v", err.Error())
		return
	}
	// parse jwt token success
	t.Logf("claims: %v", claims)
}
