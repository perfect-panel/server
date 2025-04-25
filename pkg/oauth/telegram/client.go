package telegram

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// ParseAuthDataJson parses provided json content for AuthData
func ParseAuthDataJson(content []byte) (*AuthData, error) {
	data := &AuthData{}
	err := json.Unmarshal(content, data)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling error: %w", err)
	}
	return data, nil
}

// ParseAuthDataBase64 decodes provided content from base64 and parses result for AuthData
func ParseAuthDataBase64(content []byte) (*AuthData, error) {

	decodedBytes, err := base64.RawStdEncoding.DecodeString(string(content))
	if err != nil && len(decodedBytes) == 0 {
		return nil, fmt.Errorf("base64 decoding error: %w", err)
	}
	return ParseAuthDataJson(decodedBytes)
}

// ParseAndValidateBase64 parses base64 content for AuthData and validates it
func ParseAndValidateBase64(content []byte, botToken string) (*AuthData, error) {
	authData, err := ParseAuthDataBase64(content)
	if err != nil {
		return nil, err
	}
	err = authData.Validate([]byte(botToken))
	return authData, err
}

// ParseAndValidateJson parses json content for AuthData and validates it
func ParseAndValidateJson(content []byte, botToken []byte) (*AuthData, error) {
	authData, err := ParseAuthDataJson(content)
	if err != nil {
		return nil, err
	}
	err = authData.Validate(botToken)
	return authData, err
}

// GenerateTelegramOAuthURL generates a URL for Telegram OAuth
func GenerateTelegramOAuthURL(botToken, embed, redirect string) string {
	bot := strings.Split(botToken, ":")
	uri := "https://oauth.telegram.org/auth?bot_id=%s&origin=%s&embed=%s&request_access=write&return_to=%s"
	parsedURL, err := url.Parse(redirect)
	if err != nil {
		return ""
	}
	return fmt.Sprintf(uri, bot[0], fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host), embed, redirect)
}
