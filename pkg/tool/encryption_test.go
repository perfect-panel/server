package tool

import (
	"testing"
)

func TestEncodePassWord(t *testing.T) {
	t.Logf("EncodePassWord: %v", EncodePassWord("password"))
}

func TestMultiPasswordVerify(t *testing.T) {
	pwd := "$2y$10$WFO17pdtohfeBILjEChoGeVxpDG.u9kVCKhjDAeEeNmCjIlj3tDRy"
	status := MultiPasswordVerify("bcrypt", "", "admin1", pwd)
	t.Logf("MultiPasswordVerify: %v", status)
}
