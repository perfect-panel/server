package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// getAuthDataCheckString returns a string ready to calculate validation hash.
// Ref: https://core.telegram.org/widgets/login#checking-authorization
func getAuthDataCheckString(data *AuthData) string {
	t := reflect.TypeOf(data).Elem()
	v := reflect.ValueOf(data).Elem()
	fields := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		tag, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		if v.Field(i).IsNil() || tag == "hash" {
			continue
		}
		val := v.Field(i).Elem().Interface()
		if val == nil || val == "null" {
			continue
		}
		fields = append(fields, fmt.Sprintf("%s=%v", tag, val))
	}
	sort.Strings(fields)
	return strings.Join(fields, "\n")
}

// computeHash returns a hash calculated for AuthData
// Ref: https://core.telegram.org/widgets/login#checking-authorization
func computeHash(data *AuthData, botToken []byte) string {
	checkString := getAuthDataCheckString(data)
	key := sha256.Sum256(botToken)
	h := hmac.New(sha256.New, key[:])
	h.Write([]byte(checkString))
	return hex.EncodeToString(h.Sum(nil))
}
