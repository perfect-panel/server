package tool

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/pkg/errors"

	"golang.org/x/crypto/curve25519"
)

func Curve25519Genkey(StdEncoding bool, inputBase64 string) (public, private string, err error) {
	var privateKey, publicKey []byte
	var encoding *base64.Encoding
	if StdEncoding {
		encoding = base64.StdEncoding
	} else {
		encoding = base64.RawURLEncoding
	}

	if len(inputBase64) > 0 {
		privateKey, err = encoding.DecodeString(inputBase64)
		if err != nil {
			goto out
		}
		if len(privateKey) != curve25519.ScalarSize {
			err = errors.New("Invalid length of private key.")
			goto out
		}
	}

	if privateKey == nil {
		privateKey = make([]byte, curve25519.ScalarSize)
		if _, err = rand.Read(privateKey); err != nil {
			goto out
		}
	}

	// Modify random bytes using algorithm described at:
	// https://cr.yp.to/ecdh.html.
	privateKey[0] &= 248
	privateKey[31] &= 127 | 64

	if publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint); err != nil {
		goto out
	}
	public = encoding.EncodeToString(publicKey)
	private = encoding.EncodeToString(privateKey)
out:
	return public, private, err
}
