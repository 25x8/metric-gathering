package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// CalculateHash вычисляет HMAC-SHA256 хеш от данных с использованием ключа.
// Принимает массив байтов data и строку key в качестве аргументов.
// Возвращает хеш в виде шестнадцатеричной строки.
// Используется для подписи метрик при передаче данных между агентом и сервером.
func CalculateHash(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
