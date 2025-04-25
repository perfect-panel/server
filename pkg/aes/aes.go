package pkgaes

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/forgoer/openssl"
)

// Encrypt 传入 []byte，返回 []byte 类型的加密数据
func Encrypt(plainText []byte, keyStr string) (string, string, error) {
	//get time
	nonce := fmt.Sprintf("%x", time.Now().UnixNano())
	key := generateKey(keyStr)
	iv := generateIv(nonce, keyStr)
	dst, err := openssl.AesCBCEncrypt(plainText, key, iv, openssl.PKCS7_PADDING)
	// 返回加密后的数据（包括 IV）
	return base64.StdEncoding.EncodeToString(dst), nonce, err
}

// Decrypt 传入 []byte 类型的加密数据，返回解密后的 []byte 明文数据
func Decrypt(cipherText string, keyStr string, ivStr string) (string, error) {
	decode, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	key := generateKey(keyStr)
	iv := generateIv(ivStr, keyStr)
	dst, err := openssl.AesCBCDecrypt(decode, key, iv, openssl.PKCS7_PADDING)
	return string(dst), err
}

// 生成密钥（哈希处理后保持为固定大小）
func generateKey(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:32] // AES-256 需要 32 字节密钥
}

func generateIv(iv, key string) []byte {
	h := md5.New()
	h.Write([]byte(iv))
	return generateKey(hex.EncodeToString(h.Sum(nil)) + key)
}
