package tool

import (
	"testing"
)

func TestGenerateCipher(t *testing.T) {
	pwd := GenerateCipher("serverKey", 128)
	t.Logf("pwd: %s", pwd)
}
