package tool

import (
	"crypto/sha1"
	"fmt"
)

func GenerateShortID(privateKey string) string {
	hash := sha1.New()
	hash.Write([]byte(privateKey))
	hashValue := hash.Sum(nil)
	hashString := fmt.Sprintf("%x", hashValue)
	return hashString[:8]
}
