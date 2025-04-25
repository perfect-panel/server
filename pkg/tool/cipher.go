package tool

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateCipher 根据公钥生成固定长度密文
func GenerateCipher(serverKey string, length int) string {
	h := hmac.New(sha256.New, []byte(serverKey))
	hash := h.Sum(nil)
	hashStr := hex.EncodeToString(hash)
	// Prevent overflow
	if length > len(hashStr) {
		length = len(hashStr)
	}
	return hashStr[:length]
}
