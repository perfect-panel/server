package tool

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/rand"
)

func FixedUniqueString(s string, length int, alphabet string) (string, error) {
	if alphabet == "" {
		alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	if length <= 0 {
		return "", errors.New("length must be > 0")
	}
	if length > len(alphabet) {
		return "", errors.New("length greater than available unique characters")
	}

	// Generate deterministic seed from SHA256
	hash := sha256.Sum256([]byte(s))
	seed := int64(binary.LittleEndian.Uint64(hash[:8])) // 前 8 字节

	r := rand.New(rand.NewSource(seed))

	// Copy alphabet to mutable array
	data := []rune(alphabet)

	// Deterministic shuffle (Fisher–Yates)
	for i := len(data) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		data[i], data[j] = data[j], data[i]
	}

	// Take first N characters
	return string(data[:length]), nil
}
