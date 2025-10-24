package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func CreateHash(key string, value []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(value)
	return hex.EncodeToString(h.Sum(nil))
}

func CheckHash(key, verifiableРash string, value []byte) bool {
	if verifiableРash == "" && key == "" {
		return true
	}
	return hmac.Equal([]byte(CreateHash(key, value)), []byte(verifiableРash))
}
