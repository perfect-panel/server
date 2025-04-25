package pkgaes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAes(t *testing.T) {
	params := map[string]interface{}{
		"method":   "email",
		"account":  "admin@ppanel.dev",
		"password": "password",
	}
	marshal, _ := json.Marshal(params)
	jsonStr := string(marshal)
	encrypt, iv, err := Encrypt([]byte(jsonStr), "123456")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	decrypt, err := Decrypt(encrypt, "123456", iv)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	assert.Equal(t, jsonStr, decrypt, "decrypt failed")

}
