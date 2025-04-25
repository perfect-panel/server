package turnstile

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
)

type service struct {
	timeout time.Duration
	secret  string
	url     string
}

func newService(config Config) Service {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	return &service{
		secret:  config.Secret,
		timeout: config.Timeout,
		url:     "https://challenges.cloudflare.com/turnstile/v0/siteverify",
	}
}

func (s *service) Verify(ctx context.Context, token string, ip string) (bool, error) {
	return s.verify(ctx, s.secret, token, ip, "")
}

func (s *service) VerifyIdempotent(ctx context.Context, token string, ip string, key string) (bool, error) {
	return s.verify(ctx, s.secret, token, ip, key)
}

func (s *service) RandomUUID() string {
	uuid := make([]byte, 16)
	_, _ = rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func (s *service) verify(ctx context.Context, secret string, token string, ip string, key string) (bool, error) {
	_, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("secret", secret)
	_ = writer.WriteField("response", token)
	_ = writer.WriteField("remoteip", ip)
	if key != "" {
		_ = writer.WriteField("idempotency_key", key)
	}
	_ = writer.Close()
	client := &http.Client{}
	req, _ := http.NewRequest("POST", s.url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	firstResult, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer firstResult.Body.Close()
	firstOutcome := make(map[string]interface{})
	err = json.NewDecoder(firstResult.Body).Decode(&firstOutcome)
	if err != nil {
		return false, err
	}
	if success, ok := firstOutcome["success"].(bool); ok && success {
		return true, nil
	}
	return false, nil
}
