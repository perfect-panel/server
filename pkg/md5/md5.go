package md5

import (
	"crypto/md5"
	"encoding/hex"
)

func Sign(content string) string {
	h := md5.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}
