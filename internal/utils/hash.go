package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// вычисляем HMAC-SHA256 хеш от данных с использованием ключа
func CalculateHash(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
