package tool

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/google/uuid"
)

// GenerateDeviceToken returns a 32-byte random token in URL-safe base64 (no padding).
// Used as the per-device subscription URL token (V4.3 决策 36).
func GenerateDeviceToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand should never fail on supported platforms; fall back to uuid.
		return uuid.NewString()
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// GenerateUUIDv4 returns a fresh v4 UUID string. Used for per-device proto-agnostic identity.
func GenerateUUIDv4() string {
	return uuid.NewString()
}

// DerivePasswordFromUUID returns sha256(uuid)[:16] as a hex string,
// used as the SS-protocol password derived from a device UUID (V4.3 决策 17).
func DerivePasswordFromUUID(deviceUUID string) string {
	sum := sha256.Sum256([]byte(deviceUUID))
	return hex.EncodeToString(sum[:8]) // 8 bytes hex => 16 chars
}
