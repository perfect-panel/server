package random

import (
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	chars62 = "E7gLp4jWS6kPv5DzxaY1o9sNcFmBAlUut0ZOhKVM38bqHRJfCwdrTni2QIeXGy"
	base62  = int64(len(chars62))

	chars36 = "6W1HLYPUSJ745ZAKMBQEN9DF8OVGITX320RC"
	base36  = int64(len(chars36))
)

func EncodeBase62(id int64) string {
	if id == 0 {
		return string(chars62[0])
	}

	encoded := ""
	for id > 0 {
		remainder := id % base62
		encoded = string(chars62[remainder]) + encoded
		id /= base62
	}

	index := len(chars62) - 1
	for len(encoded) < 6 {
		encoded = string(chars62[index]) + encoded
		index -= 3
		if index < 0 {
			index = len(chars62) - 1
		}
	}
	// if len(encoded) > 7 {
	// 	encoded = encoded[:7]
	// }

	return encoded
}

// EncodeBase36 ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789
func EncodeBase36(id int64) string {
	if id == 0 {
		return string(chars36[0])
	}

	encoded := ""
	for id > 0 {
		remainder := id % base36
		encoded = string(chars36[remainder]) + encoded
		id /= base36
	}

	index := len(chars36) - 1
	for len(encoded) < 6 {
		encoded = string(chars62[index]) + encoded
		index -= 3
		if index < 0 {
			index = len(chars62) - 1
		}
	}
	// if len(encoded) > 7 {
	// 	encoded = encoded[:7]
	// }

	return encoded
}

func Key(length int, keyType int) string {
	randomString := "0123456789"
	if keyType == 1 {
		randomString = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	}
	var res []byte
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		n := rand.Intn(len(randomString))
		res = append(res, randomString[n])
	}
	return string(res)
}

func KeyNew(length int, keyType int) string {
	randomString := "0123456789"
	if keyType == 1 {
		randomString = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	} else if keyType == 2 {
		randomString = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}
	var res []byte
	for i := 0; i < length; i++ {
		n := rand.Intn(len(randomString))
		res = append(res, randomString[n])
	}
	return string(res)
}

func StrToDashedString(strNum string) string {
	var result strings.Builder

	for i, ch := range strNum {
		result.WriteRune(ch)
		if (i+1)%4 == 0 && i != len(strNum)-1 {
			result.WriteRune('-')
		}
	}

	return result.String()
}
