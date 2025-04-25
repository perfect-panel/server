package tool

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/anaskhan96/go-password-encoder"
)

var options = &password.Options{SaltLen: 16, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}

func EncodePassWord(str string) string {
	salt, encodedPwd := password.Encode(str, options)
	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	return newPassword
}

func VerifyPassWord(passwd, EncodePasswd string) bool {
	info := strings.Split(EncodePasswd, "$")
	return password.Verify(passwd, info[2], info[3], options)
}

func Md5Encode(str string, isUpper bool) string {
	sum := md5.Sum([]byte(str))
	res := hex.EncodeToString(sum[:])
	//转大写，strings.ToUpper(res)
	if isUpper {
		res = strings.ToUpper(res)
	}
	return res
}
