package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// Option jwt additional data
type Option struct {
	Key string
	Val any
}

// WithOption returns Option with key-value pairs
func WithOption(key string, val any) Option {
	return Option{
		Key: key,
		Val: val,
	}
}

// NewJwtToken Generate and return jwt token with given data.
func NewJwtToken(secretKey string, iat, seconds int64, opt ...Option) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat

	for _, v := range opt {
		claims[v.Key] = v.Val
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

// ParseJwtToken Parse jwt token and return claims.
func ParseJwtToken(tokenString, secretKey string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidId
	}
	return claims, nil
}
