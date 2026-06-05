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

func TestMultiPasswordVerifySha256Salt(t *testing.T) {
	// sha256("123456" + "ppanel")
	hash := "4fb4d5ec8ec384d63cfe1faf2d9610140b310f68fd72eb0df90d3027b702b35f"
	if !MultiPasswordVerify("sha256salt", "ppanel", "123456", hash) {
		t.Fatal("sha256salt: correct password should verify")
	}
	if MultiPasswordVerify("sha256salt", "ppanel", "wrong", hash) {
		t.Fatal("sha256salt: wrong password must not verify")
	}
}
