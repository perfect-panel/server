package tool

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/anaskhan96/go-password-encoder"
	"golang.org/x/crypto/bcrypt"
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

func MultiPasswordVerify(algo, salt, password, hash string) bool {
	switch algo {
	case "md5":
		sum := md5.Sum([]byte(password))
		return hex.EncodeToString(sum[:]) == hash
	case "sha256":
		sum := sha256.Sum256([]byte(password))
		return hex.EncodeToString(sum[:]) == hash
	case "md5salt":
		sum := md5.Sum([]byte(password + salt))
		return hex.EncodeToString(sum[:]) == hash
	case "default": // PPanel's default algorithm
		return VerifyPassWord(password, hash)
	case "bcrypt":
		// Bcrypt (corresponding to PHP's password_hash/password_verify)
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		return err == nil
	}
	return false
}
